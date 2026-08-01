package admin

import (
	"fmt"
	"strings"
	"time"
)

// SetClientExpiration sets an optional UTC expiration instant for a client.
// Passing "never" removes an existing expiration.
func (s *Service) SetClientExpiration(name, value string) (resultErr error) {
	if !s.IsRoot() {
		return fmt.Errorf("root privileges required to manage clients")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("client name is required")
	}

	var expiresAt *time.Time
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("expiration is required; use RFC3339 or never")
	}
	if !strings.EqualFold(value, "never") {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return fmt.Errorf("invalid expiration %q: use RFC3339, for example 2026-08-31T23:59:59Z: %w", value, err)
		}
		parsed = parsed.UTC()
		if !parsed.After(time.Now().UTC()) {
			return fmt.Errorf("expiration must be in the future")
		}
		expiresAt = &parsed
	}

	stor, err := s.openStorage()
	if err != nil {
		return err
	}
	defer closeWithError(stor.DB, &resultErr)
	return stor.SetClientExpiration(name, expiresAt)
}
