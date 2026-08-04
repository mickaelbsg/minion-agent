package bootstrap

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const DefaultCredentialsPath = "/var/lib/minion/bootstrap-credentials.txt"

var ErrAlreadyConsumed = errors.New("bootstrap credentials are unavailable or already consumed")

// PublicationError reports a failure after the official destination was
// created. Callers must not delete the matching client automatically because
// the credential may already be recoverable from disk.
type PublicationError struct {
	Err error
}

func (e *PublicationError) Error() string { return e.Err.Error() }
func (e *PublicationError) Unwrap() error { return e.Err }

func WasPublished(err error) bool {
	var publicationErr *PublicationError
	return errors.As(err, &publicationErr)
}

// WriteCredentials stores a newly generated bootstrap credential without
// exposing it through stdout. The destination is created once, root-only, and
// published atomically only after the complete payload has reached disk.
func WriteCredentials(path, clientName, allowedIPs, apiKey string) error {
	if strings.TrimSpace(path) == "" {
		path = DefaultCredentialsPath
	}
	if strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("bootstrap API key is required")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create bootstrap credentials directory: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("secure bootstrap credentials directory: %w", err)
	}
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("bootstrap credentials already exist; consume or remove them before creating another client")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect bootstrap credentials destination: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".bootstrap-credentials-*")
	if err != nil {
		return fmt.Errorf("create temporary bootstrap credentials: %w", err)
	}
	tmpPath := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if err := tmp.Chmod(0o600); err != nil {
		cleanup()
		return fmt.Errorf("secure temporary bootstrap credentials: %w", err)
	}
	payload := fmt.Sprintf("Client: %s\nAllowed IPs: %s\nAPI Key: %s\n", clientName, allowedIPs, apiKey)
	if _, err := io.WriteString(tmp, payload); err != nil {
		cleanup()
		return fmt.Errorf("write temporary bootstrap credentials: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync temporary bootstrap credentials: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close temporary bootstrap credentials: %w", err)
	}

	// A hard link publishes the completed file atomically and refuses to
	// overwrite an existing credential file.
	if err := os.Link(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("publish bootstrap credentials: %w", err)
	}
	if err := syncDirectory(dir); err != nil {
		_ = os.Remove(tmpPath)
		return &PublicationError{Err: fmt.Errorf("sync published bootstrap credentials directory: %w", err)}
	}

	// The official destination is durable at this point. Failure to remove the
	// temporary hard link must not make setup delete the matching client.
	if err := os.Remove(tmpPath); err == nil {
		_ = syncDirectory(dir)
	}
	return nil
}

func syncDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

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
