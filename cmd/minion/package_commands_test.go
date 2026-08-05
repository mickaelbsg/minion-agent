package main

import (
	"fmt"
	"minion/internal/admin"
	"minion/internal/config"
	"minion/internal/storage"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageClientExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "minion.db")
	configPath := filepath.Join(dir, "config.json")
	configBody := fmt.Sprintf(`{
  "api": {"bind": "127.0.0.1:9870", "allow_insecure_http": false},
  "security": {"allowed_fail2ban_jails": []},
  "db_path": %q,
  "clients": []
}`, dbPath)
	if err := os.WriteFile(configPath, []byte(configBody), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	stor, err := storage.New(dbPath)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}
	if err := stor.InsertClient("bootstrap", "127.0.0.1/32", "argon2id-test-hash"); err != nil {
		_ = stor.DB.Close()
		t.Fatalf("insert client: %v", err)
	}
	if err := stor.DB.Close(); err != nil {
		t.Fatalf("close storage: %v", err)
	}

	service := admin.NewService(configPath)
	exists, err := packageClientExists(service, "bootstrap")
	if err != nil {
		t.Fatalf("inspect existing client: %v", err)
	}
	if !exists {
		t.Fatal("expected bootstrap client to exist")
	}

	exists, err = packageClientExists(service, "missing")
	if err != nil {
		t.Fatalf("inspect missing client: %v", err)
	}
	if exists {
		t.Fatal("unexpected missing client")
	}
}

func TestPackageReadinessEndpoint(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.API.Bind = "0.0.0.0:9870"
	endpoint, err := packageReadinessEndpoint(cfg)
	if err != nil {
		t.Fatalf("build endpoint: %v", err)
	}
	if endpoint != "https://127.0.0.1:9870/api/v1/health" {
		t.Fatalf("unexpected endpoint %q", endpoint)
	}

	cfg.API.Bind = "[::]:9871"
	endpoint, err = packageReadinessEndpoint(cfg)
	if err != nil {
		t.Fatalf("build IPv6 endpoint: %v", err)
	}
	if endpoint != "https://[::1]:9871/api/v1/health" {
		t.Fatalf("unexpected IPv6 endpoint %q", endpoint)
	}
}

func TestPackageReadyWithClient(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/health" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(strings.Repeat("x", 2048)))
	}))
	defer server.Close()

	if err := packageReadyWithClient(server.Client(), server.URL+"/api/v1/health"); err != nil {
		t.Fatalf("readiness failed: %v", err)
	}
}

func TestPackageReadyRejectsUnhealthyStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "starting", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	err := packageReadyWithClient(server.Client(), server.URL)
	if err == nil || !strings.Contains(err.Error(), "503") {
		t.Fatalf("expected unhealthy status error, got %v", err)
	}
}
