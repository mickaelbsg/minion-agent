package admin

import (
	"path/filepath"
	"strings"
	"testing"

	"minion/internal/config"
	"minion/internal/security"
)

func newRevocationTestService(t *testing.T) *Service {
	t.Helper()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	cfg := config.Default()
	cfg.DBPath = filepath.Join(tempDir, "minion.db")
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}
	service := NewService(configPath)
	service.IsRoot = func() bool { return true }
	return service
}

func TestRevokeClientInvalidatesKeyAndPreservesRecord(t *testing.T) {
	service := newRevocationTestService(t)
	created, err := service.CreateClient("automation", "192.0.2.10/32")
	if err != nil {
		t.Fatalf("CreateClient returned error: %v", err)
	}

	if err := service.RevokeClient("automation"); err != nil {
		t.Fatalf("RevokeClient returned error: %v", err)
	}

	clients, err := service.ListClients()
	if err != nil {
		t.Fatalf("ListClients returned error: %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("expected preserved client record, got %d", len(clients))
	}
	client := clients[0]
	if client.RevokedAt == nil {
		t.Fatal("expected revoked_at to be persisted")
	}
	if client.Enabled {
		t.Fatal("revoked client must be disabled")
	}
	if security.VerifyAPIKey(created.APIKey, client.APIKeyHash) {
		t.Fatal("old API key remained valid after revocation")
	}
	parts := strings.Split(client.APIKeyHash, "$")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		t.Fatal("replacement hash does not use the project's salt$argon2id format")
	}
	if client.APIKeyHash == created.APIKeyHash {
		t.Fatal("revocation preserved the previous API key hash")
	}
}

func TestRevokedClientCannotBeReenabledRotatedOrExpired(t *testing.T) {
	service := newRevocationTestService(t)
	if _, err := service.CreateClient("automation", "127.0.0.1/32"); err != nil {
		t.Fatalf("CreateClient returned error: %v", err)
	}
	if err := service.RevokeClient("automation"); err != nil {
		t.Fatalf("RevokeClient returned error: %v", err)
	}

	if err := service.SetClientEnabled("automation", true); err == nil {
		t.Fatal("expected enabling revoked client to fail")
	}
	if _, err := service.RotateClientAPIKey("automation"); err == nil {
		t.Fatal("expected rotating revoked client to fail")
	}
	if err := service.SetClientExpiration("automation", "never"); err == nil {
		t.Fatal("expected changing expiration of revoked client to fail")
	}
	if err := service.RevokeClient("automation"); err == nil {
		t.Fatal("expected repeated revocation to fail")
	}
}

func TestRevokeClientRequiresRootAndExistingClient(t *testing.T) {
	service := newRevocationTestService(t)
	if err := service.RevokeClient("missing"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got %v", err)
	}

	service.IsRoot = func() bool { return false }
	if err := service.RevokeClient("automation"); err == nil || !strings.Contains(err.Error(), "root") {
		t.Fatalf("expected root error, got %v", err)
	}
}
