package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Client struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	AllowedIPs []string `json:"allowed_ips"`
	APIKeyHash string   `json:"api_key_hash"`
	Enabled    bool     `json:"enabled"`
}

type Storage struct {
	DB *sql.DB
}

func New(path string) (*Storage, error) {
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
	createAuditIndexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit(timestamp);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_client_name ON audit(client_name);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_path ON audit(path);`,
	}
	for _, stmt := range createAuditIndexes {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
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

func (s *Storage) GetClients() ([]Client, error) {
	rows, err := s.DB.Query("SELECT id, name, allowed_ips, api_key_hash, enabled FROM clients")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []Client
	for rows.Next() {
		var c Client
		var ips string
		if err := rows.Scan(&c.ID, &c.Name, &ips, &c.APIKeyHash, &c.Enabled); err != nil {
			return nil, err
		}
		c.AllowedIPs = strings.Split(ips, ",")
		clients = append(clients, c)
	}
	return clients, nil
}

func (s *Storage) InsertClient(name, ips, hash string) error {
	_, err := s.DB.Exec(
		"INSERT INTO clients (name, allowed_ips, api_key_hash, enabled) VALUES (?, ?, ?, ?)",
		name, ips, hash, 1,
	)
	return err
}

func (s *Storage) UpdateClientStatus(name string, enabled bool) error {
	val := 0
	if enabled {
		val = 1
	}
	_, err := s.DB.Exec("UPDATE clients SET enabled = ? WHERE name = ?", val, name)
	return err
}

func (s *Storage) DeleteClient(name string) error {
	_, err := s.DB.Exec("DELETE FROM clients WHERE name = ?", name)
	return err
}
