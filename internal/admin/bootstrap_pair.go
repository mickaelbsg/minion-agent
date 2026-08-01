package admin

import (
	"fmt"
	"strings"
)

// PairBootstrap restricts the bootstrap client to the supplied Automation IP/CIDR.
func (s *Service) PairBootstrap(ips string) (resultErr error) {
	if !s.IsRoot() {
		return fmt.Errorf("root privileges required to pair bootstrap client")
	}
	ips = strings.TrimSpace(ips)
	if err := validateAllowedIPs(ips); err != nil {
		return err
	}

	stor, err := s.openStorage()
	if err != nil {
		return err
	}
	defer closeWithError(stor.DB, &resultErr)

	if err := stor.UpdateClientAllowedIPs("bootstrap", ips); err != nil {
		return fmt.Errorf("update bootstrap client allowlist: %w", err)
	}
	return nil
}
