package config

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
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
        // If the config file does not exist we generate a minimal one with sane
        // defaults and persist it so the binary can be started without manual
        // preparation.
        if os.IsNotExist(err) {
            cfg := Config{}
            // Apply the same defaults the original implementation used.
            cfg.API.Bind = "0.0.0.0:9870"
            cfg.DBPath = "/opt/minion/minion.db"
            cfg.Clients = []Client{}

            // Ensure the target directory exists before writing the file.
            if dir := filepath.Dir(path); dir != "." && dir != "" {
                if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
                    return nil, fmt.Errorf("failed to create config directory %s: %w", dir, mkErr)
                }
            }
            out, _ := json.MarshalIndent(&cfg, "", "  ")
            if wErr := os.WriteFile(path, out, 0o644); wErr != nil {
                return nil, fmt.Errorf("failed to write default config: %w", wErr)
            }
            return &cfg, nil
        }
        return nil, err
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    // Preserve backward‑compatibility: fill defaults when fields are omitted.
    if cfg.API.Bind == "" {
        cfg.API.Bind = "0.0.0.0:9870"
    }
    if cfg.DBPath == "" {
        cfg.DBPath = "/opt/minion/minion.db"
    }

    return &cfg, nil
}
