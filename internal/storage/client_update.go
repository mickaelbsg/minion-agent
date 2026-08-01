package storage

import "fmt"

// UpdateClientAllowedIPs replaces the allowlist for an existing client.
func (s *Storage) UpdateClientAllowedIPs(name, ips string) error {
	result, err := s.DB.Exec("UPDATE clients SET allowed_ips = ? WHERE name = ?", ips, name)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("client %q not found", name)
	}
	return nil
}
