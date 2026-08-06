package main

import (
	"fmt"
	"minion/internal/admin"
	"minion/internal/storage"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageClientStateSatisfiesInstallWithExistingNonBootstrapClient(t *testing.T) {
	t.Parallel()

	service := packageClientStateTestService(t, "automation")
	satisfied, err := packageClientStateSatisfiesInstall(service, "bootstrap")
	if err != nil {
		t.Fatalf("inspect package client state: %v", err)
	}
	if !satisfied {
		t.Fatal("existing non-bootstrap client should prevent bootstrap regeneration")
	}
}

func TestPackageClientStateSatisfiesInstallRequiresBootstrapOnEmptyDatabase(t *testing.T) {
	t.Parallel()

	service := packageClientStateTestService(t, "")
	satisfied, err := packageClientStateSatisfiesInstall(service, "bootstrap")
	if err != nil {
		t.Fatalf("inspect empty package client state: %v", err)
	}
	if satisfied {
		t.Fatal("empty database must require bootstrap creation")
	}
}

func TestPackageClientStateSatisfiesInstallDoesNotFallbackForOtherNames(t *testing.T) {
	t.Parallel()

	service := packageClientStateTestService(t, "automation")
	satisfied, err := packageClientStateSatisfiesInstall(service, "missing")
	if err != nil {
		t.Fatalf("inspect named client state: %v", err)
	}
	if satisfied {
		t.Fatal("non-bootstrap client lookup must remain exact")
	}
}

func packageClientStateTestService(t *testing.T, clientName string) *admin.Service {
	t.Helper()

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
	if clientName != "" {
		if err := stor.InsertClient(clientName, "192.0.2.10/32", "argon2id-test-hash"); err != nil {
			_ = stor.DB.Close()
			t.Fatalf("insert client: %v", err)
		}
	}
	if err := stor.DB.Close(); err != nil {
		t.Fatalf("close storage: %v", err)
	}
	return admin.NewService(configPath)
}
