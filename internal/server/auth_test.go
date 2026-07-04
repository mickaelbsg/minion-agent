package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"minion/internal/config"
	"minion/internal/security"
	"minion/internal/storage"
)

func TestAuthPrefersDatabaseClientsWhenAvailable(t *testing.T) {
	db, err := storage.New(":memory:")
	if err != nil {
		t.Fatalf("storage init error: %v", err)
	}
	defer db.DB.Close()

	dbKey := "db-secret"
	if err := db.InsertClient("db-client", "127.0.0.1/32", security.HashAPIKey(dbKey)); err != nil {
		t.Fatalf("insert db client: %v", err)
	}

	cfg := &config.Config{}
	cfg.Clients = []config.Client{
		{
			Name:       "config-client",
			AllowedIPs: []string{"127.0.0.1/32"},
			APIKeyHash: security.HashAPIKey("config-secret"),
			Enabled:    true,
		},
	}

	srv := New(cfg, db)
	handler := srv.auth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("Authorization", "Bearer config-secret")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected database-backed auth to reject config key when DB clients exist, got %d", rec.Code)
	}
}

func TestAuthFallsBackToConfigClientsWhenDatabaseIsEmpty(t *testing.T) {
	db, err := storage.New(":memory:")
	if err != nil {
		t.Fatalf("storage init error: %v", err)
	}
	defer db.DB.Close()

	cfgKey := "config-secret"
	cfg := &config.Config{}
	cfg.Clients = []config.Client{
		{
			Name:       "config-client",
			AllowedIPs: []string{"127.0.0.1/32"},
			APIKeyHash: security.HashAPIKey(cfgKey),
			Enabled:    true,
		},
	}

	srv := New(cfg, db)
	handler := srv.auth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("Authorization", "Bearer "+cfgKey)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected config-backed auth to succeed when DB has no clients, got %d", rec.Code)
	}
}
