package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type MinionConfig struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	APIKey   string `json:"api_key"`
	Insecure bool   `json:"insecure"`
}

type Config struct {
	Minions []MinionConfig `json:"minions"`
	Server  struct {
		Port int `json:"port"`
	} `json:"server"`
}

type MinionData struct {
	Name   string          `json:"name"`
	Host   string          `json:"host"`
	Online bool            `json:"online"`
	Agent  json.RawMessage `json:"agent,omitempty"`
	System json.RawMessage `json:"system,omitempty"`
	Memory json.RawMessage `json:"memory,omitempty"`
	Disk   json.RawMessage `json:"disk,omitempty"`
	Users  json.RawMessage `json:"users,omitempty"`
	Error  string          `json:"error,omitempty"`
}

type LXCContainer struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	IP      string `json:"ip"`
	Image   string `json:"image"`
	MinionIP string `json:"minion_ip,omitempty"`
}

type LXCRequest struct {
	Name  string `json:"name"`
	Image string `json:"image,omitempty"`
	Port  int    `json:"port,omitempty"`
}

type DeployRequest struct {
	ContainerName string `json:"container_name"`
	PackagePath   string `json:"package_path,omitempty"`
}

type TestResult struct {
	Test   string `json:"test"`
	Status string `json:"status"`
	Output string `json:"output,omitempty"`
}

type Server struct {
	config      Config
	lxcMu       sync.Mutex
	containerCache map[string]*LXCContainer
}

func main() {
	cfg, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	s := &Server{
		config:         cfg,
		containerCache: make(map[string]*LXCContainer),
	}

	http.HandleFunc("/", s.handleIndex)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/api/minions", s.handleMinions)
	http.HandleFunc("/api/minion/", s.handleMinionDetail)
	http.HandleFunc("/api/lxc/list", s.handleLXCList)
	http.HandleFunc("/api/lxc/create", s.handleLXCCreate)
	http.HandleFunc("/api/lxc/destroy", s.handleLXCDestroy)
	http.HandleFunc("/api/lxc/deploy", s.handleLXCDeploy)
	http.HandleFunc("/api/lxc/test", s.handleLXCTest)
	http.HandleFunc("/api/lxc/status/", s.handleLXCStatus)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Uzinha server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func loadConfig(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("reading config: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	return cfg, nil
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

func (s *Server) handleMinions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	results := make([]MinionData, 0, len(s.config.Minions))
	for _, m := range s.config.Minions {
		results = append(results, fetchMinionData(m))
	}
	json.NewEncoder(w).Encode(results)
}

func (s *Server) handleMinionDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, `{"error":"missing name parameter"}`, http.StatusBadRequest)
		return
	}
	for _, m := range s.config.Minions {
		if m.Name == name {
			json.NewEncoder(w).Encode(fetchMinionFullData(m))
			return
		}
	}
	http.Error(w, `{"error":"minion not found"}`, http.StatusNotFound)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (s *Server) handleLXCList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	containers, err := listLXCContainers()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(containers)
}

func (s *Server) handleLXCCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req LXCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		req.Name = fmt.Sprintf("minion-lab-%d", time.Now().Unix()%10000)
	}
	if req.Image == "" {
		req.Image = "debian:12"
	}

	container, err := createLXCContainer(req.Name, req.Image)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(container)
}

func (s *Server) handleLXCDestroy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := destroyLXCContainer(req.Name); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "destroyed"})
}

func (s *Server) handleLXCDeploy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	result, err := deployMinionToContainer(req.ContainerName, req.PackagePath)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleLXCTest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	results, err := testMinionInContainer(req.Name)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(results)
}

func (s *Server) handleLXCStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	name := strings.TrimPrefix(r.URL.Path, "/api/lxc/status/")
	if name == "" {
		http.Error(w, `{"error":"missing container name"}`, http.StatusBadRequest)
		return
	}

	status, err := getLXCContainerStatus(name)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(status)
}

func minionHTTPClient(insecure bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		},
		Timeout: 5 * time.Second,
	}
}

func minionFullHTTPClient(insecure bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		},
		Timeout: 10 * time.Second,
	}
}

func fetchMinionData(m MinionConfig) MinionData {
	data := MinionData{Name: m.Name, Host: m.Host}
	client := minionHTTPClient(m.Insecure)
	resp, err := client.Get(m.Host + "/api/v1/agent")
	if err != nil {
		data.Error = fmt.Sprintf("connection failed: %v", err)
		return data
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		data.Error = fmt.Sprintf("read failed: %v", err)
		return data
	}
	data.Online = resp.StatusCode == http.StatusOK
	data.Agent = json.RawMessage(body)
	return data
}

func fetchMinionFullData(m MinionConfig) MinionData {
	data := MinionData{Name: m.Name, Host: m.Host}
	client := minionFullHTTPClient(m.Insecure)
	endpoints := map[string]*json.RawMessage{
		"/api/v1/agent":  &data.Agent,
		"/api/v1/system": &data.System,
		"/api/v1/memory": &data.Memory,
		"/api/v1/disk":   &data.Disk,
		"/api/v1/users":  &data.Users,
	}
	for path, target := range endpoints {
		resp, err := client.Get(m.Host + path)
		if err != nil {
			if data.Error == "" {
				data.Error = fmt.Sprintf("endpoint %s failed: %v", path, err)
			}
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		if resp.StatusCode == http.StatusOK {
			*target = json.RawMessage(body)
			if !data.Online {
				data.Online = true
			}
		}
	}
	return data
}

func runCommand(name string, args ...string) (string, error) {
	shellArgs := make([]string, 0, len(args)+2)
	shellArgs = append(shellArgs, name)
	for _, a := range args {
		shellArgs = append(shellArgs, shellQuote(a))
	}
	cmd := exec.Command("wsl", "-e", "bash", "-c", strings.Join(shellArgs, " "))
	output, err := cmd.Output()
	if err != nil {
		stderr := ""
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		return stripControl(string(output)), fmt.Errorf("%s", stripControl(stderr))
	}
	return stripControl(string(output)), nil
}

func stripControl(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, s)
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " \t\n\"'\\$`!#&|;(){}[]<>*?~") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func runLXCCommand(args ...string) (string, error) {
	return runCommand("incus", args...)
}

func listLXCContainers() ([]LXCContainer, error) {
	output, err := runLXCCommand("list", "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}

	var rawContainers []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
		State  struct {
			Network struct {
				Eth0 struct {
					Addresses []struct {
						Address string `json:"address"`
						Family  string `json:"family"`
					} `json:"addresses"`
				} `json:"eth0"`
			} `json:"network"`
		} `json:"state"`
	}

	if err := json.Unmarshal([]byte(output), &rawContainers); err != nil {
		return nil, fmt.Errorf("parsing container list: %w", err)
	}

	containers := make([]LXCContainer, 0, len(rawContainers))
	for _, rc := range rawContainers {
		container := LXCContainer{
			Name:   rc.Name,
			Status: rc.Status,
			Image:  "debian:12",
		}
		for _, addr := range rc.State.Network.Eth0.Addresses {
			if addr.Family == "inet" {
				container.IP = addr.Address
				break
			}
		}
		containers = append(containers, container)
	}

	return containers, nil
}

func createLXCContainer(name, image string) (*LXCContainer, error) {
	ensureNAT()

	_, err := runLXCCommand("launch", "ea8c12769f00", name, "-c", "security.nesting=true")
	if err != nil {
		return nil, fmt.Errorf("creating container: %w", err)
	}

	time.Sleep(3 * time.Second)

	_, err = runLXCCommand("exec", name, "--", "systemctl", "is-system-running", "--wait")
	if err != nil {
		log.Printf("Warning: systemd wait failed: %v", err)
	}

	_, err = runLXCCommand("exec", name, "--", "apt-get", "update", "-qq")
	if err != nil {
		return nil, fmt.Errorf("updating packages: %w", err)
	}

	_, err = runLXCCommand("exec", name, "--", "apt-get", "install", "-y", "-qq",
		"fail2ban", "iptables", "openssl", "sqlite3", "curl", "gnupg")
	if err != nil {
		return nil, fmt.Errorf("installing dependencies: %w", err)
	}

	return getLXCContainerStatus(name)
}

func ensureNAT() {
	_, _ = runCommand("bash", "-c",
		"sudo sysctl -w net.ipv4.ip_forward=1 2>/dev/null; "+
			"sudo iptables -t nat -C POSTROUTING -s 10.162.89.0/24 ! -o incusbr0 -j MASQUERADE 2>/dev/null || "+
			"sudo iptables -t nat -A POSTROUTING -s 10.162.89.0/24 ! -o incusbr0 -j MASQUERADE; "+
			"sudo iptables -C FORWARD -i incusbr0 -j ACCEPT 2>/dev/null || "+
			"sudo iptables -A FORWARD -i incusbr0 -j ACCEPT; "+
			"sudo iptables -C FORWARD -o incusbr0 -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || "+
			"sudo iptables -A FORWARD -o incusbr0 -m state --state RELATED,ESTABLISHED -j ACCEPT")
}

func destroyLXCContainer(name string) error {
	_, _ = runLXCCommand("stop", name, "--force")
	_, err := runLXCCommand("delete", name, "--force")
	if err != nil {
		return fmt.Errorf("destroying container: %w", err)
	}
	return nil
}

func deployMinionToContainer(containerName, packagePath string) (map[string]interface{}, error) {
	if packagePath == "" {
		debFiles, _ := filepath.Glob("../minion_*_amd64.deb")
		if len(debFiles) > 0 {
			packagePath = latestDeb(debFiles)
		}
		if packagePath == "" || strings.Contains(packagePath, "*") {
			return nil, fmt.Errorf("no .deb package found in project root")
		}
	}

	absPath, err := filepath.Abs(packagePath)
	if err != nil {
		return nil, fmt.Errorf("resolving package path: %w", err)
	}
	wslPath := windowsToWSLPath(absPath)

	_, err = runLXCCommand("file", "push", wslPath, containerName+"/tmp/minion.deb")
	if err != nil {
		return nil, fmt.Errorf("pushing package: %w", err)
	}

	output, err := runLXCCommand("exec", containerName, "--", "bash", "-c",
		"DEBIAN_FRONTEND=noninteractive apt-get install -y -qq /tmp/minion.deb")
	if err != nil {
		return nil, fmt.Errorf("installing package: %s: %w", output, err)
	}

	status, err := getLXCContainerStatus(containerName)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"status":    "installed",
		"container": status,
		"output":    output,
	}

	bootstrap, err := runLXCCommand("exec", containerName, "--", "cat", "/var/lib/minion/bootstrap-credentials.txt")
	if err == nil {
		result["bootstrap"] = strings.TrimSpace(bootstrap)
	}

	return result, nil
}

func windowsToWSLPath(winPath string) string {
	winPath = strings.ReplaceAll(winPath, "\\", "/")
	if len(winPath) >= 2 && winPath[1] == ':' {
		drive := strings.ToLower(string(winPath[0]))
		winPath = winPath[2:]
		return "/mnt/" + drive + winPath
	}
	return winPath
}

func latestDeb(files []string) string {
	if len(files) == 0 {
		return ""
	}
	latest := files[0]
	for _, f := range files[1:] {
		if extractDebVersion(f) > extractDebVersion(latest) {
			latest = f
		}
	}
	return latest
}

func extractDebVersion(path string) string {
	base := filepath.Base(path)
	base = strings.TrimPrefix(base, "minion_")
	idx := strings.LastIndex(base, "_")
	if idx > 0 {
		return base[:idx]
	}
	return base
}

func testMinionInContainer(containerName string) ([]TestResult, error) {
	var results []TestResult

	output, err := runLXCCommand("exec", containerName, "--", "systemctl", "is-active", "minion.service")
	status := "fail"
	if err == nil && strings.TrimSpace(output) == "active" {
		status = "pass"
	}
	results = append(results, TestResult{Test: "service_active", Status: status, Output: strings.TrimSpace(output)})

	output, err = runLXCCommand("exec", containerName, "--", "curl", "--silent", "--show-error", "--fail", "--insecure",
		"https://127.0.0.1:9870/api/v1/health")
	status = "fail"
	if err == nil {
		status = "pass"
	}
	results = append(results, TestResult{Test: "health_check", Status: status, Output: strings.TrimSpace(output)})

	output, err = runLXCCommand("exec", containerName, "--", "stat", "-c", "%a", "/etc/minion/config.json")
	status = "fail"
	if err == nil && strings.TrimSpace(output) == "600" {
		status = "pass"
	}
	results = append(results, TestResult{Test: "config_permissions", Status: status, Output: strings.TrimSpace(output)})

	output, err = runLXCCommand("exec", containerName, "--", "stat", "-c", "%a", "/opt/minion/minion.db")
	status = "fail"
	if err == nil && strings.TrimSpace(output) == "600" {
		status = "pass"
	}
	results = append(results, TestResult{Test: "db_permissions", Status: status, Output: strings.TrimSpace(output)})

	return results, nil
}

func getLXCContainerStatus(name string) (*LXCContainer, error) {
	output, err := runLXCCommand("list", "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}

	var rawContainers []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
		State  struct {
			Network struct {
				Eth0 struct {
					Addresses []struct {
						Address string `json:"address"`
						Family  string `json:"family"`
					} `json:"addresses"`
				} `json:"eth0"`
			} `json:"network"`
		} `json:"state"`
	}

	if err := json.Unmarshal([]byte(output), &rawContainers); err != nil {
		return nil, fmt.Errorf("parsing container list: %w", err)
	}

	for _, rc := range rawContainers {
		if rc.Name == name {
			container := &LXCContainer{
				Name:   rc.Name,
				Status: rc.Status,
				Image:  "debian:12",
			}
			for _, addr := range rc.State.Network.Eth0.Addresses {
				if addr.Family == "inet" {
					container.IP = addr.Address
					break
				}
			}
			return container, nil
		}
	}

	return nil, fmt.Errorf("container %s not found", name)
}
