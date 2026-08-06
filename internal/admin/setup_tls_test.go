package admin

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"minion/internal/config"
)

type rejectingCommandRunner struct {
	calls int
}

func (r *rejectingCommandRunner) Run(name string, args ...string) ([]byte, error) {
	r.calls++
	return nil, fmt.Errorf("external command %q is not available", name)
}

func TestSetupGeneratesTLSWithoutExternalCommands(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	dbPath := filepath.Join(tempDir, "minion.db")
	tlsDir := filepath.Join(tempDir, "tls")

	cfg := config.Default()
	cfg.DBPath = dbPath
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	runner := &rejectingCommandRunner{}
	service := NewService(configPath)
	service.TLSDir = tlsDir
	service.ServiceUnit = ""
	service.CommandRunner = runner
	service.IsRoot = func() bool { return true }

	result, err := service.Setup(SetupOptions{
		ClientName: "bootstrap",
		ClientIPs:  "127.0.0.1/32",
	})
	if err != nil {
		t.Fatalf("Setup returned error without OpenSSL: %v", err)
	}
	if runner.calls != 0 {
		t.Fatalf("setup invoked %d external commands", runner.calls)
	}
	if !result.CertGenerated {
		t.Fatal("expected setup to generate TLS assets")
	}
	if !result.BootstrapCreated || result.APIKey == "" {
		t.Fatal("expected setup to create bootstrap credentials")
	}
	if _, err := tls.LoadX509KeyPair(result.TLSCertPath, result.TLSKeyPath); err != nil {
		t.Fatalf("load generated TLS pair: %v", err)
	}

	keyInfo, err := os.Stat(result.TLSKeyPath)
	if err != nil {
		t.Fatalf("stat TLS key: %v", err)
	}
	if keyInfo.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected TLS key mode: %o", keyInfo.Mode().Perm())
	}
	dirInfo, err := os.Stat(tlsDir)
	if err != nil {
		t.Fatalf("stat TLS directory: %v", err)
	}
	if dirInfo.Mode().Perm() != 0o700 {
		t.Fatalf("unexpected TLS directory mode: %o", dirInfo.Mode().Perm())
	}
}
