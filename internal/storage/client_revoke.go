package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// RevokeClient permanently invalidates a client credential while preserving
// the record for audit. The replacement hash must belong to a random secret
// that is discarded by the caller.
func (s *Storage) RevokeClient(name, replacementHash string, revokedAt time.Time) error {
	result, err := s.DB.Exec(
		`UPDATE clients
		 SET api_key_hash = ?, enabled = 0, revoked_at = ?
		 WHERE name = ? AND revoked_at IS NULL`,
		replacementHash,
		revokedAt.UTC().Format(time.RFC3339),
		name,
	)
	if err != nil {
		return fmt.Errorf("failed to revoke client: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to verify client revocation: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
