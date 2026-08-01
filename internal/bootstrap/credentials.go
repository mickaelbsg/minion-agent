package bootstrap

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const DefaultCredentialsPath = "/var/lib/minion/bootstrap-credentials.txt"

var ErrAlreadyConsumed = errors.New("bootstrap credentials are unavailable or already consumed")

// Consume prints bootstrap credentials once and removes the source file only
// after the complete payload has been written successfully.
func Consume(path string, output io.Writer, isRoot func() bool) error {
	if isRoot == nil || !isRoot() {
		return fmt.Errorf("root privileges required to read bootstrap credentials")
	}
	if strings.TrimSpace(path) == "" {
		path = DefaultCredentialsPath
	}
	if output == nil {
		return fmt.Errorf("output writer is required")
	}

	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrAlreadyConsumed
		}
		return fmt.Errorf("inspect bootstrap credentials: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("bootstrap credentials file must not be a symbolic link")
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("bootstrap credentials path is not a regular file")
	}
	if info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("bootstrap credentials permissions are insecure: expected 0600 or stricter")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read bootstrap credentials: %w", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return fmt.Errorf("bootstrap credentials file is empty")
	}

	payload := data
	if payload[len(payload)-1] != '\n' {
		payload = append(payload, '\n')
	}
	if _, err := output.Write(payload); err != nil {
		return fmt.Errorf("write bootstrap credentials: %w", err)
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove consumed bootstrap credentials: %w", err)
	}
	return nil
}
