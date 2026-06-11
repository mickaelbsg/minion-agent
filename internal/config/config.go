package config

import (
	"encoding/json"
	"os"
)

type Client struct {
	Name       string   `json:"name"`
	AllowedIPs []string `json:"allowed_ips"`
	APIKeyHash string   `json:"api_key_hash"`
	Enabled    bool     `json:"enabled"`
}

type Config struct {
	API struct {
		Bind string `json:"bind"`
	} `json:"api"`
	Clients []Client `json:"clients"`
	DBPath  string   `json:"db_path"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.API.Bind == "" {
		cfg.API.Bind = "0.0.0.0:9870"
	}

	if cfg.DBPath == "" {
		cfg.DBPath = "/opt/minion/minion.db"
	}

	return &cfg, nil
}
