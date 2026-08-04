package admin

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"minion/internal/config"
	"minion/internal/security"
	"minion/internal/storage"
)

const (
	defaultTLSDir      = "/etc/minion/tls"
	defaultServiceUnit = "minion.service"
)

type CommandRunner interface {
	Run(name string, args ...string) ([]byte, error)
}

type execRunner struct{}

func (execRunner) Run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

type Service struct {
	ConfigPath    string
	TLSDir        string
	ServiceUnit   string
	CommandRunner CommandRunner
	IsRoot        func() bool
}

type Status struct {
	ConfigPath    string
	ConfigExists  bool
	DBPath        string
	DBExists      bool
	TLSCertPath   string
	TLSKeyPath    string
	TLSCertExists bool
	TLSKeyExists  bool
	ClientCount   int
	ServiceStatus string
}

type SetupOptions struct {
	ClientName string
	ClientIPs  string
}

type SetupResult struct {
	ConfigPath       string
	DBPath           string
	TLSCertPath      string
	TLSKeyPath       string
	CertGenerated    bool
	ServiceStarted   bool
	BootstrapCreated bool
	ClientName       string
	ClientIPs        string
	APIKey           string
}

type ConfigUpdate struct {
	Bind              string
	DBPath            string
	AllowInsecureHTTP bool
}

type CreatedClient struct {
	Name       string
	AllowedIPs string
	APIKey     string
	APIKeyHash string
}

func NewService(configPath string) *Service {
	return &Service{
		ConfigPath:    configPath,
		TLSDir:        defaultTLSDir,
		ServiceUnit:   defaultServiceUnit,
		CommandRunner: execRunner{},
		IsRoot: func() bool {
			return os.Geteuid() == 0
		},
	}
}

func (s *Service) ReadConfig() (*config.Config, error) {
	return config.Read(s.ConfigPath)
}

func (s *Service) ReadConfigOrDefault() (*config.Config, error) {
	cfg, err := s.ReadConfig()
	if err == nil {
		return cfg, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return config.Default(), nil
	}
	return nil, err
}

func (s *Service) SaveConfig(update ConfigUpdate) error {
	if !s.IsRoot() {
		return fmt.Errorf("root privileges required to update config")
	}
	if err := validateBind(update.Bind); err != nil {
		return err
	}

	cfg, err := s.ReadConfigOrDefault()
	if err != nil {
		return err
	}
	cfg.API.Bind = update.Bind
	cfg.DBPath = update.DBPath
	cfg.API.AllowInsecureHTTP = update.AllowInsecureHTTP

	return config.Save(s.ConfigPath, cfg)
}

func (s *Service) InspectStatus() (Status, error) {
	status := Status{
		ConfigPath:    s.ConfigPath,
		TLSCertPath:   filepath.Join(s.TLSDir, "minion.crt"),
		TLSKeyPath:    filepath.Join(s.TLSDir, "minion.key"),
		ServiceStatus: "unknown",
	}

	if _, err := os.Stat(s.ConfigPath); err == nil {
		status.ConfigExists = true
	}

	cfg, err := s.ReadConfig()
	switch {
	case err == nil:
		status.DBPath = cfg.DBPath
	case errors.Is(err, os.ErrNotExist):
		status.DBPath = config.Default().DBPath
	default:
		return status, err
	}

	if _, err := os.Stat(status.DBPath); err == nil {
		status.DBExists = true
	}
	if _, err := os.Stat(status.TLSCertPath); err == nil {
		status.TLSCertExists = true
	}
	if _, err := os.Stat(status.TLSKeyPath); err == nil {
		status.TLSKeyExists = true
	}

	if status.DBExists {
		stor, err := storage.New(status.DBPath)
		if err == nil {
			defer stor.DB.Close()
			clients, listErr := stor.GetClients()
			if listErr == nil {
				status.ClientCount = len(clients)
			}
		}
	}

	if s.ServiceUnit != "" {
		out, runErr := s.CommandRunner.Run("systemctl", "is-active", s.ServiceUnit)
		if runErr == nil {
			status.ServiceStatus = strings.TrimSpace(string(out))
		} else {
			trimmed := strings.TrimSpace(string(out))
			if trimmed != "" {
				status.ServiceStatus = trimmed
			}
		}
	}

	return status, nil
}

func (s *Service) Setup(opts SetupOptions) (result SetupResult, resultErr error) {
	if !s.IsRoot() {
		return SetupResult{}, fmt.Errorf("setup requires root privileges")
	}

	if strings.TrimSpace(opts.ClientName) == "" {
		opts.ClientName = "default"
	}
	if strings.TrimSpace(opts.ClientIPs) == "" {
		opts.ClientIPs = "127.0.0.1/32"
	}
	if err := validateAllowedIPs(opts.ClientIPs); err != nil {
		return SetupResult{}, err
	}

	result = SetupResult{
		ConfigPath:  s.ConfigPath,
		TLSCertPath: filepath.Join(s.TLSDir, "minion.crt"),
		TLSKeyPath:  filepath.Join(s.TLSDir, "minion.key"),
		ClientName:  opts.ClientName,
		ClientIPs:   opts.ClientIPs,
	}

	if err := os.MkdirAll(s.TLSDir, 0o755); err != nil {
		return SetupResult{}, fmt.Errorf("failed to create TLS directory: %w", err)
	}

	if _, err := os.Stat(result.TLSCertPath); errors.Is(err, os.ErrNotExist) {
		if _, err := s.CommandRunner.Run(
			"openssl", "req", "-newkey", "rsa:2048", "-nodes",
			"-keyout", result.TLSKeyPath, "-x509", "-days", "365", "-out", result.TLSCertPath,
			"-subj", "/CN=minion",
		); err != nil {
			return SetupResult{}, fmt.Errorf("failed to generate TLS cert: %w", err)
		}
		result.CertGenerated = true
	}

	cfg, err := config.Load(s.ConfigPath)
	if err != nil {
		return SetupResult{}, fmt.Errorf("failed to load config: %w", err)
	}
	result.DBPath = cfg.DBPath

	if err := os.Chmod(s.ConfigPath, 0o600); err != nil {
		return SetupResult{}, fmt.Errorf("failed to set config permissions: %w", err)
	}

	stor, err := storage.New(cfg.DBPath)
	if err != nil {
		return SetupResult{}, fmt.Errorf("failed to initialise storage: %w", err)
	}
	defer closeWithError(stor.DB, &resultErr)

	if err := os.Chmod(cfg.DBPath, 0o600); err != nil {
		return SetupResult{}, fmt.Errorf("failed to set database permissions: %w", err)
	}

	clients, err := stor.GetClients()
	if err != nil {
		return SetupResult{}, fmt.Errorf("failed to list clients: %w", err)
	}

	if len(clients) == 0 {
		key, err := security.GenerateAPIKey()
		if err != nil {
			return SetupResult{}, fmt.Errorf("failed to generate API key: %w", err)
		}
		hash, err := security.HashAPIKeyWithError(key)
		if err != nil {
			return SetupResult{}, fmt.Errorf("failed to hash API key: %w", err)
		}
		if err := stor.InsertClient(opts.ClientName, opts.ClientIPs, hash); err != nil {
			return SetupResult{}, fmt.Errorf("failed to create bootstrap client: %w", err)
		}
		result.APIKey = key
		result.BootstrapCreated = true
	}

	if s.ServiceUnit != "" {
		if _, err := s.CommandRunner.Run("systemctl", "enable", "--now", s.ServiceUnit); err == nil {
			result.ServiceStarted = true
		}
	}

	return result, nil
}

func (s *Service) ListClients() (clients []storage.Client, resultErr error) {
	stor, err := s.openStorage()
	if err != nil {
		return nil, err
	}
	defer closeWithError(stor.DB, &resultErr)
	return stor.GetClients()
}

func (s *Service) CreateClient(name, ips string) (created CreatedClient, resultErr error) {
	if !s.IsRoot() {
		return CreatedClient{}, fmt.Errorf("root privileges required to manage clients")
	}
	if strings.TrimSpace(name) == "" {
		return CreatedClient{}, fmt.Errorf("client name is required")
	}
	if err := validateAllowedIPs(ips); err != nil {
		return CreatedClient{}, err
	}

	stor, err := s.openStorage()
	if err != nil {
		return CreatedClient{}, err
	}
	defer closeWithError(stor.DB, &resultErr)

	key, err := security.GenerateAPIKey()
	if err != nil {
		return CreatedClient{}, fmt.Errorf("failed to generate API key: %w", err)
	}
	hash, err := security.HashAPIKeyWithError(key)
	if err != nil {
		return CreatedClient{}, fmt.Errorf("failed to hash API key: %w", err)
	}
	if err := stor.InsertClient(name, ips, hash); err != nil {
		return CreatedClient{}, err
	}

	created = CreatedClient{
		Name:       name,
		AllowedIPs: ips,
		APIKey:     key,
		APIKeyHash: hash,
	}
	return created, nil
}

func (s *Service) SetClientEnabled(name string, enabled bool) (resultErr error) {
	if !s.IsRoot() {
		return fmt.Errorf("root privileges required to manage clients")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("client name is required")
	}

	stor, err := s.openStorage()
	if err != nil {
		return err
	}
	defer closeWithError(stor.DB, &resultErr)

	return stor.UpdateClientStatus(name, enabled)
}

func (s *Service) DeleteClient(name string) (resultErr error) {
	if !s.IsRoot() {
		return fmt.Errorf("root privileges required to manage clients")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("client name is required")
	}

	stor, err := s.openStorage()
	if err != nil {
		return err
	}
	defer closeWithError(stor.DB, &resultErr)

	return stor.DeleteClient(name)
}

func (s *Service) openStorage() (*storage.Storage, error) {
	cfg, err := config.Load(s.ConfigPath)
	switch {
	case err == nil:
		return storage.New(cfg.DBPath)
	case errors.Is(err, os.ErrNotExist):
		return storage.New("/etc/minion/minion.db")
	default:
		return nil, err
	}
}

func validateBind(bind string) error {
	if strings.TrimSpace(bind) == "" {
		return fmt.Errorf("bind address is required")
	}
	if _, err := net.ResolveTCPAddr("tcp", bind); err != nil {
		return fmt.Errorf("invalid bind address: %w", err)
	}
	return nil
}

func validateAllowedIPs(ips string) error {
	parts := strings.Split(ips, ",")
	if len(parts) == 0 {
		return fmt.Errorf("at least one IP or CIDR is required")
	}

	valid := 0
	for _, part := range parts {
		entry := strings.TrimSpace(part)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			if _, _, err := net.ParseCIDR(entry); err != nil {
				return fmt.Errorf("invalid CIDR %q", entry)
			}
		} else if net.ParseIP(entry) == nil {
			return fmt.Errorf("invalid IP %q", entry)
		}
		valid++
	}

	if valid == 0 {
		return fmt.Errorf("at least one IP or CIDR is required")
	}

	return nil
}
