package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

type Server struct {
	config Config
}

func main() {
	cfg, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	s := &Server{config: cfg}

	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/api/minions", s.handleMinions)
	http.HandleFunc("/api/minion/", s.handleMinionDetail)

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

func minionHTTPClient(insecure bool) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}
	return &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}
}

func minionFullHTTPClient(insecure bool) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}
	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
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
		"/api/v1/agent":   &data.Agent,
		"/api/v1/system":  &data.System,
		"/api/v1/memory":  &data.Memory,
		"/api/v1/disk":    &data.Disk,
		"/api/v1/users":   &data.Users,
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
