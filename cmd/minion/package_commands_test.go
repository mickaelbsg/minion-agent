package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
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
	"time"
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

func TestPackageEnsureTLSCreatesValidSecurePair(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "tls")
	certPath := filepath.Join(dir, "minion.crt")
	keyPath := filepath.Join(dir, "minion.key")
	now := time.Date(2026, 8, 6, 0, 0, 0, 0, time.UTC)

	created, err := packageEnsureTLS(certPath, keyPath, now)
	if err != nil {
		t.Fatalf("ensure TLS: %v", err)
	}
	if !created {
		t.Fatal("expected TLS pair to be created")
	}
	if _, err := tls.LoadX509KeyPair(certPath, keyPath); err != nil {
		t.Fatalf("load generated pair: %v", err)
	}
	if mode := fileMode(t, keyPath); mode != 0o600 {
		t.Fatalf("private key mode = %o, expected 600", mode)
	}
	if mode := fileMode(t, certPath); mode != 0o644 {
		t.Fatalf("certificate mode = %o, expected 644", mode)
	}
	if mode := fileMode(t, dir); mode != 0o700 {
		t.Fatalf("TLS directory mode = %o, expected 700", mode)
	}

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("read certificate: %v", err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("generated certificate is not PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	if cert.Subject.CommonName != "minion" {
		t.Fatalf("unexpected common name %q", cert.Subject.CommonName)
	}
	if cert.NotAfter.Sub(cert.NotBefore) < 364*24*time.Hour {
		t.Fatalf("unexpected certificate validity: %s", cert.NotAfter.Sub(cert.NotBefore))
	}
}

func TestPackageEnsureTLSPreservesExistingPair(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "tls")
	certPath := filepath.Join(dir, "minion.crt")
	keyPath := filepath.Join(dir, "minion.key")
	created, err := packageEnsureTLS(certPath, keyPath, time.Now())
	if err != nil || !created {
		t.Fatalf("initial ensure TLS: created=%v err=%v", created, err)
	}
	certBefore, _ := os.ReadFile(certPath)
	keyBefore, _ := os.ReadFile(keyPath)

	created, err = packageEnsureTLS(certPath, keyPath, time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("preserve TLS: %v", err)
	}
	if created {
		t.Fatal("existing TLS pair was unexpectedly regenerated")
	}
	certAfter, _ := os.ReadFile(certPath)
	keyAfter, _ := os.ReadFile(keyPath)
	if string(certAfter) != string(certBefore) || string(keyAfter) != string(keyBefore) {
		t.Fatal("existing TLS material changed")
	}
}

func TestPackageEnsureTLSRejectsIncompletePair(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "tls")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	certPath := filepath.Join(dir, "minion.crt")
	keyPath := filepath.Join(dir, "minion.key")
	original := []byte("do-not-overwrite")
	if err := os.WriteFile(keyPath, original, 0o600); err != nil {
		t.Fatalf("write partial key: %v", err)
	}

	created, err := packageEnsureTLS(certPath, keyPath, time.Now())
	if err == nil || !strings.Contains(err.Error(), "incomplete TLS pair") {
		t.Fatalf("expected incomplete pair error, got created=%v err=%v", created, err)
	}
	keyAfter, readErr := os.ReadFile(keyPath)
	if readErr != nil {
		t.Fatalf("read preserved key: %v", readErr)
	}
	if string(keyAfter) != string(original) {
		t.Fatal("partial existing key was overwritten")
	}
	if _, statErr := os.Stat(certPath); !os.IsNotExist(statErr) {
		t.Fatalf("certificate unexpectedly created: %v", statErr)
	}
}

func TestPackageEnsureTLSRejectsSymlink(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "tls")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, []byte("target"), 0o600); err != nil {
		t.Fatalf("write target: %v", err)
	}
	keyPath := filepath.Join(dir, "minion.key")
	if err := os.Symlink(target, keyPath); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	_, err := packageEnsureTLS(filepath.Join(dir, "minion.crt"), keyPath, time.Now())
	if err == nil || !strings.Contains(err.Error(), "unsafe TLS path") {
		t.Fatalf("expected unsafe path error, got %v", err)
	}
}

func fileMode(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	return info.Mode().Perm()
}

func TestPackageReadinessEndpoint(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.API.Bind = "0.0.0.0:9870"
	endpoint, err := packageReadinessEndpoint(cfg, true)
	if err != nil {
		t.Fatalf("build endpoint: %v", err)
	}
	if endpoint != "https://127.0.0.1:9870/api/v1/health" {
		t.Fatalf("unexpected endpoint %q", endpoint)
	}

	cfg.API.Bind = "[::]:9871"
	endpoint, err = packageReadinessEndpoint(cfg, true)
	if err != nil {
		t.Fatalf("build IPv6 endpoint: %v", err)
	}
	if endpoint != "https://[::1]:9871/api/v1/health" {
		t.Fatalf("unexpected IPv6 endpoint %q", endpoint)
	}
}

func TestPackageReadinessEndpointMatchesServerTLSPrecedence(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.API.Bind = "127.0.0.1:9870"
	cfg.API.AllowInsecureHTTP = true

	endpoint, err := packageReadinessEndpoint(cfg, true)
	if err != nil {
		t.Fatalf("build TLS endpoint: %v", err)
	}
	if endpoint != "https://127.0.0.1:9870/api/v1/health" {
		t.Fatalf("expected HTTPS while TLS assets exist, got %q", endpoint)
	}

	endpoint, err = packageReadinessEndpoint(cfg, false)
	if err != nil {
		t.Fatalf("build HTTP endpoint: %v", err)
	}
	if endpoint != "http://127.0.0.1:9870/api/v1/health" {
		t.Fatalf("expected HTTP only without TLS assets, got %q", endpoint)
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
