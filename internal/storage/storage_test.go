package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStorageCreatesDBDir(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nonexistent", "minion.db")
	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() {
		if err := s.DB.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	})
	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Fatalf("directory %s was not created", filepath.Dir(dbPath))
	}
}

func TestStorageRejectsInvalidDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "invalid.db")
	if err := os.WriteFile(dbPath, []byte("not a sqlite database"), 0o600); err != nil {
		t.Fatalf("failed to create invalid database: %v", err)
	}

	s, err := New(dbPath)
	if err == nil {
		if s != nil && s.DB != nil {
			_ = s.DB.Close()
		}
		t.Fatal("expected initialization error for invalid database")
	}
	if s != nil {
		t.Fatalf("expected nil storage on initialization failure, got %#v", s)
	}
}

func TestGetClientsReturnsPersistedClients(t *testing.T) {
	s, err := New(filepath.Join(t.TempDir(), "minion.db"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() {
		if err := s.DB.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	})

	if err := s.InsertClient("severino", "192.0.2.10/32,198.51.100.0/24", "hash"); err != nil {
		t.Fatalf("failed to insert client: %v", err)
	}

	clients, err := s.GetClients()
	if err != nil {
		t.Fatalf("failed to get clients: %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("expected one client, got %d", len(clients))
	}
	if clients[0].Name != "severino" {
		t.Fatalf("expected client severino, got %q", clients[0].Name)
	}
	if len(clients[0].AllowedIPs) != 2 {
		t.Fatalf("expected two allowed IP entries, got %d", len(clients[0].AllowedIPs))
	}
}
