package storage

import (
    "database/sql"
    "fmt"
    "os"
    "path/filepath"

    _ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	DB *sql.DB
}

func New(path string) (*Storage, error) {
    // Ensure the directory for the SQLite file exists. The caller may provide
    // a path like "/opt/minion/minion.db" where the parent directory does not
    // exist on a fresh system. We create the directory with permissions 0755
    // before opening the database.
    if dir := filepath.Dir(path); dir != "." && dir != "" {
        if err := os.MkdirAll(dir, 0o755); err != nil {
            return nil, fmt.Errorf("failed to create DB directory %s: %w", dir, err)
        }
    }

    db, err := sql.Open("sqlite3", path)
    if err != nil {
        return nil, err
    }

	if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		return nil, fmt.Errorf("failed to set journal mode: %w", err)
	}

	if err := initSchema(db); err != nil {
		return nil, err
	}

	return &Storage{DB: db}, nil
}

func initSchema(db *sql.DB) error {
	createClients := `CREATE TABLE IF NOT EXISTS clients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		allowed_ips TEXT NOT NULL,
		api_key_hash TEXT NOT NULL,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(createClients); err != nil {
		return err
	}

	createAudit := `CREATE TABLE IF NOT EXISTS audit (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		client_name TEXT,
		ip TEXT,
		method TEXT,
		path TEXT,
		status INTEGER
	);`

	if _, err := db.Exec(createAudit); err != nil {
		return err
	}

	return nil
}

func (s *Storage) InsertAudit(clientName, ip, method, path string, status int) error {
	if s == nil || s.DB == nil {
		return nil
	}

	_, err := s.DB.Exec(
		`INSERT INTO audit (client_name, ip, method, path, status) VALUES (?, ?, ?, ?, ?)`,
		clientName, ip, method, path, status,
	)
	return err
}
