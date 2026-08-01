package storage

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestClientExpirationDisablesExpiredClient(t *testing.T) {
	s, err := New(filepath.Join(t.TempDir(), "minion.db"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = s.DB.Close() })

	if err := s.InsertClient("automation", "127.0.0.1/32", "hash"); err != nil {
		t.Fatalf("failed to insert client: %v", err)
	}
	past := time.Now().UTC().Add(-time.Minute)
	if err := s.SetClientExpiration("automation", &past); err != nil {
		t.Fatalf("failed to set expiration: %v", err)
	}

	clients, err := s.GetClients()
	if err != nil {
		t.Fatalf("failed to list clients: %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("expected one client, got %d", len(clients))
	}
	if clients[0].Enabled {
		t.Fatal("expected expired client to be unavailable for authentication")
	}
	if clients[0].ExpiresAt == nil || !clients[0].IsExpired(time.Now().UTC()) {
		t.Fatal("expected persisted expiration to be parsed")
	}
}

func TestClientExpirationCanBeCleared(t *testing.T) {
	s, err := New(filepath.Join(t.TempDir(), "minion.db"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { _ = s.DB.Close() })

	if err := s.InsertClient("automation", "127.0.0.1/32", "hash"); err != nil {
		t.Fatalf("failed to insert client: %v", err)
	}
	future := time.Now().UTC().Add(time.Hour)
	if err := s.SetClientExpiration("automation", &future); err != nil {
		t.Fatalf("failed to set expiration: %v", err)
	}
	if err := s.SetClientExpiration("automation", nil); err != nil {
		t.Fatalf("failed to clear expiration: %v", err)
	}

	clients, err := s.GetClients()
	if err != nil {
		t.Fatalf("failed to list clients: %v", err)
	}
	if clients[0].ExpiresAt != nil {
		t.Fatal("expected expiration to be cleared")
	}
	if !clients[0].Enabled {
		t.Fatal("expected client to remain enabled")
	}
}

func TestClientSchemaMigratesExistingDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("failed to open legacy database: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE clients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		allowed_ips TEXT NOT NULL,
		api_key_hash TEXT NOT NULL,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		t.Fatalf("failed to create legacy schema: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("failed to close legacy database: %v", err)
	}

	s, err := New(path)
	if err != nil {
		t.Fatalf("expected migration to succeed: %v", err)
	}
	defer s.DB.Close()

	rows, err := s.DB.Query(`PRAGMA table_info(clients)`)
	if err != nil {
		t.Fatalf("failed to inspect schema: %v", err)
	}
	defer rows.Close()
	found := false
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, primaryKey int
		var defaultValue interface{}
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("failed to scan schema: %v", err)
		}
		if name == "expires_at" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected expires_at column after migration")
	}
}
