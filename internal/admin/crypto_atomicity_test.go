package admin

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"minion/internal/config"
	"minion/internal/storage"
)

const testPlaintextAPIKey = "minion_sk_test_plaintext_must_not_leak"

func newAtomicityTestService(t *testing.T, withClient bool) (*Service, string) {
	t.Helper()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	dbPath := filepath.Join(tempDir, "minion.db")
	cfg := config.Default()
	cfg.DBPath = dbPath
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	stor, err := storage.New(dbPath)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}
	if withClient {
		if err := stor.InsertClient("automation", "127.0.0.1/32", "original-hash"); err != nil {
			stor.DB.Close()
			t.Fatalf("insert client: %v", err)
		}
	}
	if err := stor.DB.Close(); err != nil {
		t.Fatalf("close storage: %v", err)
	}

	service := NewService(configPath)
	service.IsRoot = func() bool { return true }
	service.GenerateAPIKey = func() (string, error) { return testPlaintextAPIKey, nil }
	service.HashAPIKey = func(string) (string, error) { return "", errors.New("entropy unavailable") }
	return service, dbPath
}

func readAtomicityClient(t *testing.T, dbPath string) []storage.Client {
	t.Helper()
	stor, err := storage.New(dbPath)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}
	defer stor.DB.Close()
	clients, err := stor.GetClients()
	if err != nil {
		t.Fatalf("list clients: %v", err)
	}
	return clients
}

func assertSecretNotLeaked(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected cryptographic failure")
	}
	if strings.Contains(err.Error(), testPlaintextAPIKey) {
		t.Fatalf("error exposed plaintext API key: %v", err)
	}
}

func TestCreateClientHashFailureDoesNotInsertClient(t *testing.T) {
	service, dbPath := newAtomicityTestService(t, false)

	created, err := service.CreateClient("automation", "127.0.0.1/32")
	assertSecretNotLeaked(t, err)
	if created.APIKey != "" || created.APIKeyHash != "" {
		t.Fatalf("failed creation returned credential material: %+v", created)
	}
	if clients := readAtomicityClient(t, dbPath); len(clients) != 0 {
		t.Fatalf("expected no clients after hash failure, got %+v", clients)
	}
}

func TestRotateClientHashFailurePreservesPreviousHash(t *testing.T) {
	service, dbPath := newAtomicityTestService(t, true)

	apiKey, err := service.RotateClientAPIKey("automation")
	assertSecretNotLeaked(t, err)
	if apiKey != "" {
		t.Fatal("failed rotation returned plaintext API key")
	}

	clients := readAtomicityClient(t, dbPath)
	if len(clients) != 1 {
		t.Fatalf("expected one client, got %+v", clients)
	}
	if clients[0].APIKeyHash != "original-hash" {
		t.Fatalf("rotation failure changed stored hash to %q", clients[0].APIKeyHash)
	}
	if !clients[0].Enabled || clients[0].IsRevoked() {
		t.Fatalf("rotation failure changed client state: %+v", clients[0])
	}
}

func TestRevokeClientHashFailurePreservesClientState(t *testing.T) {
	service, dbPath := newAtomicityTestService(t, true)

	err := service.RevokeClient("automation")
	assertSecretNotLeaked(t, err)

	clients := readAtomicityClient(t, dbPath)
	if len(clients) != 1 {
		t.Fatalf("expected one client, got %+v", clients)
	}
	if clients[0].APIKeyHash != "original-hash" {
		t.Fatalf("revocation failure changed stored hash to %q", clients[0].APIKeyHash)
	}
	if !clients[0].Enabled || clients[0].IsRevoked() {
		t.Fatalf("revocation failure changed client state: %+v", clients[0])
	}
}
