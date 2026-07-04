package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCreatesDefaultConfigWithRestrictedPermissions(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "minion", "config.json")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.API.Bind != "0.0.0.0:9870" {
		t.Fatalf("unexpected bind address: %q", cfg.API.Bind)
	}

	if cfg.API.AllowInsecureHTTP {
		t.Fatal("default config must keep insecure HTTP disabled")
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}

	if perms := info.Mode().Perm(); perms != 0o600 {
		t.Fatalf("expected config permissions 0600, got %04o", perms)
	}
}
