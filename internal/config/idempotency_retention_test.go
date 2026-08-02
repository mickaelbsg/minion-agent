package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIdempotencyRetentionDefaultsForLegacyConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	legacy := []byte(`{
  "api": {"bind": "127.0.0.1:9870"},
  "security": {},
  "clients": [],
  "db_path": "/tmp/minion.db"
}`)
	if err := os.WriteFile(path, legacy, 0o600); err != nil {
		t.Fatalf("write legacy config error = %v", err)
	}

	cfg, err := Read(path)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if cfg.Security.IdempotencyRetentionHours != 168 {
		t.Fatalf("retention = %d, want 168", cfg.Security.IdempotencyRetentionHours)
	}
}

func TestIdempotencyRetentionPreservesConfiguredValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	configured := []byte(`{
  "api": {"bind": "127.0.0.1:9870"},
  "security": {"idempotency_retention_hours": 336},
  "clients": [],
  "db_path": "/tmp/minion.db"
}`)
	if err := os.WriteFile(path, configured, 0o600); err != nil {
		t.Fatalf("write config error = %v", err)
	}

	cfg, err := Read(path)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if cfg.Security.IdempotencyRetentionHours != 336 {
		t.Fatalf("retention = %d, want 336", cfg.Security.IdempotencyRetentionHours)
	}
}

func TestIdempotencyRetentionRejectsExcessiveValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	configured := []byte(`{
  "api": {"bind": "127.0.0.1:9870"},
  "security": {"idempotency_retention_hours": 2562048},
  "clients": [],
  "db_path": "/tmp/minion.db"
}`)
	if err := os.WriteFile(path, configured, 0o600); err != nil {
		t.Fatalf("write config error = %v", err)
	}

	if _, err := Read(path); err == nil {
		t.Fatal("expected excessive retention value to be rejected")
	}
}
