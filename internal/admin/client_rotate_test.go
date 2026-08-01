package admin

import (
	"path/filepath"
	"strings"
	"testing"

	"minion/internal/config"
	"minion/internal/security"
)

func TestRotateClientAPIKeyInvalidatesOldKeyAndPreservesClient(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	cfg := config.Default()
	cfg.DBPath = filepath.Join(tempDir, "minion.db")
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	service := NewService(configPath)
	service.IsRoot = func() bool { return true }

	created, err := service.CreateClient("automation", "192.0.2.10/32,198.51.100.0/24")
	if err != nil {
		t.Fatalf("CreateClient returned error: %v", err)
	}
	if err := service.SetClientEnabled("automation", false); err != nil {
		t.Fatalf("disable client: %v", err)
	}

	newKey, err := service.RotateClientAPIKey("automation")
	if err != nil {
		t.Fatalf("RotateClientAPIKey returned error: %v", err)
	}
	if newKey == "" || newKey == created.APIKey {
		t.Fatal("expected a distinct new API key")
	}

	clients, err := service.ListClients()
	if err != nil {
		t.Fatalf("ListClients returned error: %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("expected one client, got %d", len(clients))
	}

	client := clients[0]
	if client.Name != "automation" {
		t.Fatalf("client name changed: %q", client.Name)
	}
	if strings.Join(client.AllowedIPs, ",") != "192.0.2.10/32,198.51.100.0/24" {
		t.Fatalf("allowed IPs changed: %v", client.AllowedIPs)
	}
	if client.Enabled {
		t.Fatal("rotation must preserve disabled state")
	}
	if security.VerifyAPIKey(created.APIKey, client.APIKeyHash) {
		t.Fatal("old API key remained valid after rotation")
	}
	if !security.VerifyAPIKey(newKey, client.APIKeyHash) {
		t.Fatal("new API key does not match stored hash")
	}
}

func TestRotateClientAPIKeyRejectsUnknownClient(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	cfg := config.Default()
	cfg.DBPath = filepath.Join(tempDir, "minion.db")
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	service := NewService(configPath)
	service.IsRoot = func() bool { return true }

	_, err := service.RotateClientAPIKey("missing")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestRotateClientAPIKeyRequiresRoot(t *testing.T) {
	service := NewService(filepath.Join(t.TempDir(), "config.json"))
	service.IsRoot = func() bool { return false }

	_, err := service.RotateClientAPIKey("automation")
	if err == nil || !strings.Contains(err.Error(), "root") {
		t.Fatalf("expected root error, got %v", err)
	}
}
