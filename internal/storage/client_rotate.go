package storage

import (
	"database/sql"
	"fmt"
)

// UpdateClientAPIKeyHash replaces the stored API key hash for an existing
// client. It returns sql.ErrNoRows when the client does not exist so callers
// cannot report a successful rotation that changed nothing.
func (s *Storage) UpdateClientAPIKeyHash(name, hash string) error {
	result, err := s.DB.Exec(
		"UPDATE clients SET api_key_hash = ? WHERE name = ?",
		hash,
		name,
	)
	if err != nil {
		return fmt.Errorf("failed to update client API key hash: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to verify client API key rotation: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
