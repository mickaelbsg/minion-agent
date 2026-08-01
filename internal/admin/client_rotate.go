package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"minion/internal/security"
)

// RotateClientAPIKey replaces the stored hash for an existing client and
// returns the new plaintext API key exactly once to the caller.
func (s *Service) RotateClientAPIKey(name string) (apiKey string, resultErr error) {
	if !s.IsRoot() {
		return "", fmt.Errorf("root privileges required to manage clients")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("client name is required")
	}

	stor, err := s.openStorage()
	if err != nil {
		return "", err
	}
	defer closeWithError(stor.DB, &resultErr)

	key, err := security.GenerateAPIKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}

	if err := stor.UpdateClientAPIKeyHash(name, security.HashAPIKey(key)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("client %q not found", name)
		}
		return "", err
	}

	return key, nil
}
