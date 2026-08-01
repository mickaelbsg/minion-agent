package server

import (
	"io"
	"minion/internal/storage"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewHTTPServerAppliesSecurityTimeouts(t *testing.T) {
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	srv := newHTTPServer("127.0.0.1:9870", handler)

	if srv.Addr != "127.0.0.1:9870" {
		t.Fatalf("unexpected address: %s", srv.Addr)
	}
	if srv.Handler == nil {
		t.Fatal("expected handler to be configured")
	}
	if srv.ReadHeaderTimeout != readHeaderTimeout {
		t.Fatalf("unexpected ReadHeaderTimeout: %s", srv.ReadHeaderTimeout)
	}
	if srv.ReadTimeout != readTimeout {
		t.Fatalf("unexpected ReadTimeout: %s", srv.ReadTimeout)
	}
	if srv.WriteTimeout != writeTimeout {
		t.Fatalf("unexpected WriteTimeout: %s", srv.WriteTimeout)
	}
	if srv.IdleTimeout != idleTimeout {
		t.Fatalf("unexpected IdleTimeout: %s", srv.IdleTimeout)
	}
}

func TestLimitRequestBodyAcceptsValidPayload(t *testing.T) {
	srv := &Server{}
	var received string
	handler := srv.limitRequestBody(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		received = string(body)
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/test", strings.NewReader(`{"ok":true}`))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", res.Code)
	}
	if received != `{"ok":true}` {
		t.Fatalf("unexpected body: %q", received)
	}
}

func TestLimitRequestBodyRejectsDeclaredOversizedPayload(t *testing.T) {
	srv := &Server{}
	called := false
	handler := srv.limitRequestBody(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/test", strings.NewReader("x"))
	req.ContentLength = maxRequestBodyBytes + 1
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", res.Code)
	}
	if called {
		t.Fatal("next handler must not run for oversized payload")
	}
}

func TestLimitRequestBodyRejectsChunkedOversizedPayload(t *testing.T) {
	srv := &Server{}
	called := false
	handler := srv.limitRequestBody(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	body := strings.NewReader(strings.Repeat("x", maxRequestBodyBytes+1))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/test", body)
	req.ContentLength = -1
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", res.Code)
	}
	if called {
		t.Fatal("next handler must not run for oversized payload")
	}
}

func TestLimitRequestBodyAuditsRejectedPayload(t *testing.T) {
	store, err := storage.New(":memory:")
	if err != nil {
		t.Fatalf("storage init: %v", err)
	}
	defer store.DB.Close()

	srv := &Server{storage: store}
	handler := srv.limitRequestBody(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler must not run for oversized payload")
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/fail2ban/unban", strings.NewReader("x"))
	req.ContentLength = maxRequestBodyBytes + 1
	req.RemoteAddr = "192.0.2.40:43210"
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", res.Code)
	}

	var count int
	var clientName, sourceIP, action, detail string
	err = store.DB.QueryRow(`
		SELECT COUNT(*), COALESCE(client_name, ''), ip, action, detail
		FROM audit
		WHERE method = ? AND path = ? AND status = ?
	`, http.MethodPost, "/api/v1/fail2ban/unban", http.StatusRequestEntityTooLarge).Scan(
		&count,
		&clientName,
		&sourceIP,
		&action,
		&detail,
	)
	if err != nil {
		t.Fatalf("query audit: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 audit entry, got %d", count)
	}
	if clientName != "" {
		t.Fatalf("expected unauthenticated client name, got %q", clientName)
	}
	if sourceIP != "192.0.2.40" {
		t.Fatalf("unexpected source IP: %q", sourceIP)
	}
	if action != "request_body_rejected" {
		t.Fatalf("unexpected action: %q", action)
	}
	if detail != "reason=declared_size_exceeded" {
		t.Fatalf("unexpected detail: %q", detail)
	}
}
