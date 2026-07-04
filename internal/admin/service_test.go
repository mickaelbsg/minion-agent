package admin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"minion/internal/config"
	"minion/internal/storage"
)

type fakeRunner struct {
	out map[string][]byte
	err map[string]error
}

func (f fakeRunner) Run(name string, args ...string) ([]byte, error) {
	key := strings.Join(append([]string{name}, args...), " ")
	return f.out[key], f.err[key]
}

func TestSaveConfigUpdatesValues(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := config.Default()
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	service := NewService(configPath)
	service.IsRoot = func() bool { return true }

	err := service.SaveConfig(ConfigUpdate{
		Bind:              "127.0.0.1:9999",
		DBPath:            "/tmp/minion-test.db",
		AllowInsecureHTTP: true,
	})
	if err != nil {
		t.Fatalf("SaveConfig returned error: %v", err)
	}

	updated, err := config.Read(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	if updated.API.Bind != "127.0.0.1:9999" {
		t.Fatalf("unexpected bind: %q", updated.API.Bind)
	}
	if updated.DBPath != "/tmp/minion-test.db" {
		t.Fatalf("unexpected db path: %q", updated.DBPath)
	}
	if !updated.API.AllowInsecureHTTP {
		t.Fatal("expected allow_insecure_http=true")
	}
}

func TestSaveConfigRejectsInvalidBind(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	service := NewService(configPath)
	service.IsRoot = func() bool { return true }

	err := service.SaveConfig(ConfigUpdate{
		Bind:              "not-a-bind",
		DBPath:            "/tmp/minion-test.db",
		AllowInsecureHTTP: false,
	})
	if err == nil {
		t.Fatal("expected invalid bind error")
	}
}

func TestCreateListUpdateAndDeleteClients(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	dbPath := filepath.Join(tempDir, "minion.db")
	cfg := config.Default()
	cfg.DBPath = dbPath
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	service := NewService(configPath)
	service.IsRoot = func() bool { return true }

	created, err := service.CreateClient("api", "127.0.0.1/32")
	if err != nil {
		t.Fatalf("CreateClient returned error: %v", err)
	}
	if created.APIKey == "" {
		t.Fatal("expected API key to be generated")
	}

	clients, err := service.ListClients()
	if err != nil {
		t.Fatalf("ListClients returned error: %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("expected 1 client, got %d", len(clients))
	}

	if err := service.SetClientEnabled("api", false); err != nil {
		t.Fatalf("SetClientEnabled returned error: %v", err)
	}
	clients, err = service.ListClients()
	if err != nil {
		t.Fatalf("ListClients returned error: %v", err)
	}
	if clients[0].Enabled {
		t.Fatal("expected client to be disabled")
	}

	if err := service.DeleteClient("api"); err != nil {
		t.Fatalf("DeleteClient returned error: %v", err)
	}
	clients, err = service.ListClients()
	if err != nil {
		t.Fatalf("ListClients returned error: %v", err)
	}
	if len(clients) != 0 {
		t.Fatalf("expected 0 clients after delete, got %d", len(clients))
	}
}

func TestInspectStatusDetectsExistingAssets(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	dbPath := filepath.Join(tempDir, "minion.db")
	tlsDir := filepath.Join(tempDir, "tls")
	if err := os.MkdirAll(tlsDir, 0o755); err != nil {
		t.Fatalf("mkdir tls: %v", err)
	}

	cfg := config.Default()
	cfg.DBPath = dbPath
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	stor, err := storage.New(dbPath)
	if err != nil {
		t.Fatalf("storage init: %v", err)
	}
	defer stor.DB.Close()
	if err := stor.InsertClient("api", "127.0.0.1/32", "hash"); err != nil {
		t.Fatalf("insert client: %v", err)
	}

	certPath := filepath.Join(tlsDir, "minion.crt")
	keyPath := filepath.Join(tlsDir, "minion.key")
	if err := os.WriteFile(certPath, []byte("cert"), 0o600); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte("key"), 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	service := NewService(configPath)
	service.TLSDir = tlsDir
	service.CommandRunner = fakeRunner{
		out: map[string][]byte{
			"systemctl is-active minion.service": []byte("active\n"),
		},
		err: map[string]error{},
	}

	status, err := service.InspectStatus()
	if err != nil {
		t.Fatalf("InspectStatus returned error: %v", err)
	}

	if !status.ConfigExists || !status.DBExists || !status.TLSCertExists || !status.TLSKeyExists {
		t.Fatalf("expected all setup assets to exist: %+v", status)
	}
	if status.ClientCount != 1 {
		t.Fatalf("expected 1 client, got %d", status.ClientCount)
	}
	if status.ServiceStatus != "active" {
		t.Fatalf("expected active service, got %q", status.ServiceStatus)
	}
}

func TestSetupRequiresRoot(t *testing.T) {
	service := NewService(filepath.Join(t.TempDir(), "config.json"))
	service.IsRoot = func() bool { return false }

	_, err := service.Setup(SetupOptions{})
	if err == nil {
		t.Fatal("expected root error")
	}
	if !strings.Contains(err.Error(), "root") {
		t.Fatalf("expected root error, got %v", err)
	}
}
