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

// Read validates and reads bootstrap credentials without removing them.
func Read(path string, isRoot func() bool) ([]byte, error) {
	if isRoot == nil || !isRoot() {
		return nil, fmt.Errorf("root privileges required to read bootstrap credentials")
	}
	if strings.TrimSpace(path) == "" {
		path = DefaultCredentialsPath
	}

	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrAlreadyConsumed
		}
		return nil, fmt.Errorf("inspect bootstrap credentials: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("bootstrap credentials file must not be a symbolic link")
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("bootstrap credentials path is not a regular file")
	}
	if info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("bootstrap credentials permissions are insecure: expected 0600 or stricter")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read bootstrap credentials: %w", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return nil, fmt.Errorf("bootstrap credentials file is empty")
	}
	if data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	return data, nil
}

// Remove deletes bootstrap credentials after they have been delivered safely.
func Remove(path string) error {
	if strings.TrimSpace(path) == "" {
		path = DefaultCredentialsPath
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove consumed bootstrap credentials: %w", err)
	}
	return nil
}

// Consume prints bootstrap credentials once and removes the source file only
// after the complete payload has been written successfully.
func Consume(path string, output io.Writer, isRoot func() bool) error {
	if output == nil {
		return fmt.Errorf("output writer is required")
	}
	payload, err := Read(path, isRoot)
	if err != nil {
		return err
	}
	if _, err := output.Write(payload); err != nil {
		return fmt.Errorf("write bootstrap credentials: %w", err)
	}
	return Remove(path)
}
