package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var defaultAllowedFail2BanJails = []string{"sshd", "apache-auth", "recidive"}

type Client struct {
	Name       string   `json:"name"`
	AllowedIPs []string `json:"allowed_ips"`
	APIKeyHash string   `json:"api_key_hash"`
	Enabled    bool     `json:"enabled"`
}

type RateLimitConfig struct {
	IPBurst         int     `json:"ip_burst"`
	IPRefillPerSec  float64 `json:"ip_refill_per_second"`
	ClientBurst     int     `json:"client_burst"`
	ClientRefillSec float64 `json:"client_refill_per_second"`
}

type SecurityConfig struct {
	AllowedFail2BanJails []string        `json:"allowed_fail2ban_jails"`
	RateLimit            RateLimitConfig `json:"rate_limit"`
}

type Config struct {
	API struct {
		Bind              string `json:"bind"`
		AllowInsecureHTTP bool   `json:"allow_insecure_http"`
	} `json:"api"`
	Security SecurityConfig `json:"security"`
	Clients  []Client       `json:"clients"`
	DBPath   string         `json:"db_path"`
}

func Default() *Config {
	cfg := &Config{}
	cfg.API.Bind = "0.0.0.0:9870"
	cfg.DBPath = "/opt/minion/minion.db"
	cfg.Clients = []Client{}
	cfg.Security.AllowedFail2BanJails = append([]string{}, defaultAllowedFail2BanJails...)
	applyRateLimitDefaults(&cfg.Security.RateLimit)
	return cfg
}

func Read(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	applyDefaults(&cfg)

	return &cfg, nil
}

func Save(path string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config must not be nil")
	}

	applyDefaults(cfg)

	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create config directory %s: %w", dir, err)
		}
	}

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, out, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func Load(path string) (*Config, error) {
	cfg, err := Read(path)
	if err == nil {
		return cfg, nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	cfg = Default()
	if err := Save(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.API.Bind == "" {
		cfg.API.Bind = "0.0.0.0:9870"
	}
	if cfg.DBPath == "" {
		cfg.DBPath = "/opt/minion/minion.db"
	}
	if cfg.Clients == nil {
		cfg.Clients = []Client{}
	}
	if len(cfg.Security.AllowedFail2BanJails) == 0 {
		cfg.Security.AllowedFail2BanJails = append([]string{}, defaultAllowedFail2BanJails...)
	}
	applyRateLimitDefaults(&cfg.Security.RateLimit)
}

func applyRateLimitDefaults(cfg *RateLimitConfig) {
	if cfg.IPBurst <= 0 {
		cfg.IPBurst = 30
	}
	if cfg.IPRefillPerSec <= 0 {
		cfg.IPRefillPerSec = 5
	}
	if cfg.ClientBurst <= 0 {
		cfg.ClientBurst = 60
	}
	if cfg.ClientRefillSec <= 0 {
		cfg.ClientRefillSec = 10
	}
}
