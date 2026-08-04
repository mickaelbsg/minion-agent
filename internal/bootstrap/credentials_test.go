package bootstrap

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteCredentialsPublishesRootOnlyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "bootstrap.txt")
	if err := WriteCredentials(path, "bootstrap", "127.0.0.1/32", "minion_sk_secret"); err != nil {
		t.Fatalf("write credentials failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat credentials: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected credentials mode 0600, got %o", info.Mode().Perm())
	}
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat credentials directory: %v", err)
	}
	if dirInfo.Mode().Perm() != 0o700 {
		t.Fatalf("expected credentials directory mode 0700, got %o", dirInfo.Mode().Perm())
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read credentials: %v", err)
	}
	got := string(payload)
	for _, expected := range []string{"Client: bootstrap", "Allowed IPs: 127.0.0.1/32", "API Key: minion_sk_secret"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("payload missing %q: %q", expected, got)
		}
	}
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".bootstrap-credentials-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary credential files remained: %v", matches)
	}
}

func TestWriteCredentialsRefusesToOverwriteExistingSecret(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bootstrap.txt")
	if err := WriteCredentials(path, "first", "127.0.0.1/32", "first-secret"); err != nil {
		t.Fatal(err)
	}
	if err := WriteCredentials(path, "second", "192.0.2.1/32", "second-secret"); err == nil {
		t.Fatal("expected overwrite protection error")
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "second-secret") || !strings.Contains(string(payload), "first-secret") {
		t.Fatalf("existing credentials were replaced: %q", payload)
	}
}

func TestWriteCredentialsRejectsEmptyAPIKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bootstrap.txt")
	if err := WriteCredentials(path, "bootstrap", "127.0.0.1/32", ""); err == nil {
		t.Fatal("expected empty API key rejection")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("credentials file must not exist, got %v", err)
	}
}

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
