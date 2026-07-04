package server

import (
	"minion/internal/config"
	"minion/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuditMiddlewareCreatesEntry(t *testing.T) {
	// Setup in‑memory storage
	cfg := config.Config{DBPath: ":memory:"}
	// storage.New expects path string historically, but we have modified signature to accept path string only.
	// Use storage.New directly with path
	s, err := storage.New(cfg.DBPath)
	if err != nil {
		t.Fatalf("storage init error: %v", err)
	}
	defer s.DB.Close()

	// Create server with storage (assuming NewServer exists that takes storage)
	// Minimal config with bind address (not used for this test)
	srvCfg := &config.Config{}
	srv := New(srvCfg, s)
	// Simple handler returning 200
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// Wrap with audit middleware
	audited := srv.audit(handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	audited.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Result().StatusCode)
	}
	// Verify audit row exists
	row := s.DB.QueryRow("SELECT COUNT(*) FROM audit WHERE method=? AND path=?", "GET", "/test")
	var cnt int
	if err := row.Scan(&cnt); err != nil {
		t.Fatalf("query error: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 audit entry, got %d", cnt)
	}
}
