package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Client struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	AllowedIPs []string   `json:"allowed_ips"`
	APIKeyHash string     `json:"api_key_hash"`
	Enabled    bool       `json:"enabled"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

func (c Client) IsExpired(now time.Time) bool {
	return c.ExpiresAt != nil && !now.Before(*c.ExpiresAt)
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
		return nil, closeDBOnError(db, fmt.Errorf("failed to set journal mode: %w", err))
	}
	if err := initSchema(db); err != nil {
		return nil, closeDBOnError(db, fmt.Errorf("failed to initialize schema: %w", err))
	}
	return &Storage{DB: db}, nil
}

func closeDBOnError(db *sql.DB, initErr error) error {
	if closeErr := db.Close(); closeErr != nil {
		return errors.Join(initErr, fmt.Errorf("failed to close database after initialization error: %w", closeErr))
	}
	return initErr
}

func initSchema(db *sql.DB) error {
	createClients := `CREATE TABLE IF NOT EXISTS clients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		allowed_ips TEXT NOT NULL,
		api_key_hash TEXT NOT NULL,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		expires_at TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createClients); err != nil {
		return err
	}
	if err := migrateClientSchema(db); err != nil {
		return err
	}
	createAudit := `CREATE TABLE IF NOT EXISTS audit (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		client_name TEXT,
		ip TEXT,
		method TEXT,
		path TEXT,
		status INTEGER,
		action TEXT,
		target TEXT,
		detail TEXT
	);`
	if _, err := db.Exec(createAudit); err != nil {
		return err
	}
	if err := migrateAuditSchema(db); err != nil {
		return err
	}
	createAuditIndexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit(timestamp);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_client_name ON audit(client_name);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_path ON audit(path);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_action ON audit(action);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_target ON audit(target);`,
	}
	for _, stmt := range createAuditIndexes {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func migrateClientSchema(db *sql.DB) error {
	_, err := db.Exec(`ALTER TABLE clients ADD COLUMN expires_at TEXT;`)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return err
	}
	return nil
}

func migrateAuditSchema(db *sql.DB) error {
	migrations := []string{
		`ALTER TABLE audit ADD COLUMN action TEXT;`,
		`ALTER TABLE audit ADD COLUMN target TEXT;`,
		`ALTER TABLE audit ADD COLUMN detail TEXT;`,
	}
	for _, stmt := range migrations {
		if _, err := db.Exec(stmt); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
			return err
		}
	}
	return nil
}

func (s *Storage) InsertAudit(clientName, ip, method, path string, status int) error {
	return s.InsertAuditDetail(clientName, ip, method, path, status, "", "", "")
}

func (s *Storage) InsertAuditDetail(clientName, ip, method, path string, status int, action, target, detail string) error {
	if s == nil || s.DB == nil {
		return nil
	}
	_, err := s.DB.Exec(
		`INSERT INTO audit (client_name, ip, method, path, status, action, target, detail) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		clientName, ip, method, path, status, action, target, detail,
	)
	return err
}

func (s *Storage) GetClients() (clients []Client, err error) {
	rows, err := s.DB.Query("SELECT id, name, allowed_ips, api_key_hash, enabled, expires_at FROM clients")
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close client rows: %w", closeErr))
		}
	}()

	for rows.Next() {
		var c Client
		var ips string
		var expiresAt sql.NullString
		if err := rows.Scan(&c.ID, &c.Name, &ips, &c.APIKeyHash, &c.Enabled, &expiresAt); err != nil {
			return nil, err
		}
		c.AllowedIPs = strings.Split(ips, ",")
		if expiresAt.Valid && strings.TrimSpace(expiresAt.String) != "" {
			parsed, parseErr := time.Parse(time.RFC3339, expiresAt.String)
			if parseErr != nil {
				return nil, fmt.Errorf("client %q has invalid expires_at value: %w", c.Name, parseErr)
			}
			parsed = parsed.UTC()
			c.ExpiresAt = &parsed
		}
		clients = append(clients, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while iterating clients: %w", err)
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
	result, err := s.DB.Exec("UPDATE clients SET enabled = ? WHERE name = ?", val, name)
	if err != nil {
		return err
	}
	return requireAffectedClient(result, name)
}

func (s *Storage) SetClientExpiration(name string, expiresAt *time.Time) error {
	var value interface{}
	if expiresAt != nil {
		value = expiresAt.UTC().Format(time.RFC3339)
	}
	result, err := s.DB.Exec("UPDATE clients SET expires_at = ? WHERE name = ?", value, name)
	if err != nil {
		return err
	}
	return requireAffectedClient(result, name)
}

func requireAffectedClient(result sql.Result, name string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("client %q not found", name)
	}
	return nil
}

func (s *Storage) DeleteClient(name string) error {
	result, err := s.DB.Exec("DELETE FROM clients WHERE name = ?", name)
	if err != nil {
		return err
	}
	return requireAffectedClient(result, name)
}
