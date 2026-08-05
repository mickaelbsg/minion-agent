package main

import (
	"fmt"
	"minion/internal/admin"
	"minion/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageClientExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "minion.db")
	configPath := filepath.Join(dir, "config.json")
	configBody := fmt.Sprintf(`{
  "api": {"bind": "127.0.0.1:9870", "allow_insecure_http": false},
  "security": {"allowed_fail2ban_jails": []},
  "db_path": %q,
  "clients": []
}`, dbPath)
	if err := os.WriteFile(configPath, []byte(configBody), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	stor, err := storage.New(dbPath)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}
	if err := stor.InsertClient("bootstrap", "127.0.0.1/32", "argon2id-test-hash"); err != nil {
		_ = stor.DB.Close()
		t.Fatalf("insert client: %v", err)
	}
	if err := stor.DB.Close(); err != nil {
		t.Fatalf("close storage: %v", err)
	}

	service := admin.NewService(configPath)
	exists, err := packageClientExists(service, "bootstrap")
	if err != nil {
		t.Fatalf("inspect existing client: %v", err)
	}
	if !exists {
		t.Fatal("expected bootstrap client to exist")
	}

	exists, err = packageClientExists(service, "missing")
	if err != nil {
		t.Fatalf("inspect missing client: %v", err)
	}
	if exists {
		t.Fatal("unexpected missing client")
	}
}
