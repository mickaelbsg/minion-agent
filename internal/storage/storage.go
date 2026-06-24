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
	ID         int
	Name       string
	AllowedIPs []string
	APIKeyHash string
	Enabled    bool
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

// GetClients recupera todos os clientes ativos do banco de dados
func (s *Storage) GetClients() ([]Client, error) {
	rows, err := s.DB.Query("SELECT id, name, allowed_ips, api_key_hash, enabled FROM clients WHERE enabled = 1")
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

// InsertClient insere um novo cliente no banco de dados
func (s *Storage) InsertClient(name, ips, hash string) error {
	_, err := s.DB.Exec(
		"INSERT INTO clients (name, allowed_ips, api_key_hash) VALUES (?, ?, ?)",
		name, ips, hash,
	)
	return err
}
