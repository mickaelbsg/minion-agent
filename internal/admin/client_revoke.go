package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"minion/internal/security"
)

// RevokeClient permanently invalidates a client API key while preserving the
// client record for audit and incident response.
func (s *Service) RevokeClient(name string) (resultErr error) {
	if !s.IsRoot() {
		return fmt.Errorf("root privileges required to manage clients")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("client name is required")
	}

	stor, err := s.openStorage()
	if err != nil {
		return err
	}
	defer closeWithError(stor.DB, &resultErr)

	discardedSecret, err := security.GenerateAPIKey()
	if err != nil {
		return fmt.Errorf("failed to generate revocation secret: %w", err)
	}
	discardedHash, err := security.HashAPIKeyWithError(discardedSecret)
	if err != nil {
		return fmt.Errorf("failed to hash revocation secret: %w", err)
	}
	if err := stor.RevokeClient(name, discardedHash, time.Now().UTC()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("client %q not found or already revoked", name)
		}
		return err
	}
	return nil
}
