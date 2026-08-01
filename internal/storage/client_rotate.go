package storage

import (
	"database/sql"
	"fmt"
)

// UpdateClientAPIKeyHash replaces the stored API key hash for an existing,
// non-revoked client. It returns sql.ErrNoRows when the client does not exist
// or was permanently revoked.
func (s *Storage) UpdateClientAPIKeyHash(name, hash string) error {
	result, err := s.DB.Exec(
		"UPDATE clients SET api_key_hash = ? WHERE name = ? AND revoked_at IS NULL",
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
