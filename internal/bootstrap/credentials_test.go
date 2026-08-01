package bootstrap

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestConsumePrintsAndRemovesCredentials(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bootstrap.txt")
	if err := os.WriteFile(path, []byte("Client: automation\nAPI Key: secret"), 0o600); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := Consume(path, &output, func() bool { return true }); err != nil {
		t.Fatalf("consume failed: %v", err)
	}
	if got := output.String(); got != "Client: automation\nAPI Key: secret\n" {
		t.Fatalf("unexpected output %q", got)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected credentials file to be removed, got %v", err)
	}
}

func TestConsumeRequiresRoot(t *testing.T) {
	err := Consume("unused", io.Discard, func() bool { return false })
	if err == nil {
		t.Fatal("expected root validation error")
	}
}

func TestConsumeRejectsInsecurePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bootstrap.txt")
	if err := os.WriteFile(path, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Consume(path, io.Discard, func() bool { return true }); err == nil {
		t.Fatal("expected insecure permission error")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("credentials must remain after rejection: %v", err)
	}
}

func TestConsumeRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "bootstrap.txt")
	if err := os.WriteFile(target, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if err := Consume(link, io.Discard, func() bool { return true }); err == nil {
		t.Fatal("expected symlink rejection")
	}
}

func TestConsumePreservesFileWhenOutputFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bootstrap.txt")
	if err := os.WriteFile(path, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := Consume(path, failingWriter{}, func() bool { return true })
	if err == nil {
		t.Fatal("expected output error")
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("credentials must remain when output fails: %v", statErr)
	}
}

func TestConsumeReturnsAlreadyConsumed(t *testing.T) {
	err := Consume(filepath.Join(t.TempDir(), "missing"), io.Discard, func() bool { return true })
	if !errors.Is(err, ErrAlreadyConsumed) {
		t.Fatalf("expected ErrAlreadyConsumed, got %v", err)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("output failed")
}
