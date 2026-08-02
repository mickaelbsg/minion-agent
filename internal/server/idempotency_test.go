package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"minion/internal/storage"
)

func newIdempotencyTestServer(t *testing.T) *Server {
	t.Helper()
	store, err := storage.New(filepath.Join(t.TempDir(), "minion.db"))
	if err != nil {
		t.Fatalf("storage.New() error = %v", err)
	}
	t.Cleanup(func() { _ = store.DB.Close() })
	return &Server{storage: store}
}

func idempotencyRequest(body, requestID, client string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/fail2ban/unban", strings.NewReader(body))
	r.Header.Set("X-Request-ID", requestID)
	return withClientName(r, client)
}

func TestIdempotentActionReplaysCompletedResponse(t *testing.T) {
	s := newIdempotencyTestServer(t)
	calls := 0
	handler := s.idempotentAction("fail2ban_unban", func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"unbanned"}`))
	})

	first := httptest.NewRecorder()
	handler(first, idempotencyRequest(`{"ip":"192.0.2.10","jail":"sshd"}`, "req-00000001", "automation"))
	if first.Code != http.StatusOK {
		t.Fatalf("first status = %d, want 200; body=%s", first.Code, first.Body.String())
	}

	second := httptest.NewRecorder()
	handler(second, idempotencyRequest(`{"ip":"192.0.2.10","jail":"sshd"}`, "req-00000001", "automation"))
	if second.Code != http.StatusOK {
		t.Fatalf("replay status = %d, want 200; body=%s", second.Code, second.Body.String())
	}
	if calls != 1 {
		t.Fatalf("handler calls = %d, want 1", calls)
	}
	if second.Body.String() != first.Body.String() {
		t.Fatalf("replay body = %q, want %q", second.Body.String(), first.Body.String())
	}
	if got := second.Header().Get("X-Request-ID"); got != "req-00000001" {
		t.Fatalf("X-Request-ID = %q", got)
	}
}

func TestIdempotentActionRejectsMissingAndReusedRequestID(t *testing.T) {
	s := newIdempotencyTestServer(t)
	calls := 0
	handler := s.idempotentAction("fail2ban_unban", func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	})

	missing := httptest.NewRecorder()
	handler(missing, idempotencyRequest(`{"ip":"192.0.2.10"}`, "", "automation"))
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("missing request id status = %d, want 400", missing.Code)
	}

	first := httptest.NewRecorder()
	handler(first, idempotencyRequest(`{"ip":"192.0.2.10"}`, "req-00000002", "automation"))
	if first.Code != http.StatusNoContent {
		t.Fatalf("first status = %d, want 204", first.Code)
	}

	conflict := httptest.NewRecorder()
	handler(conflict, idempotencyRequest(`{"ip":"192.0.2.11"}`, "req-00000002", "automation"))
	if conflict.Code != http.StatusConflict {
		t.Fatalf("payload conflict status = %d, want 409; body=%s", conflict.Code, conflict.Body.String())
	}
	if calls != 1 {
		t.Fatalf("handler calls = %d, want 1", calls)
	}
}

func TestIdempotentActionDoesNotClaimRequestIDForUnsupportedMethod(t *testing.T) {
	s := newIdempotencyTestServer(t)
	calls := 0
	handler := s.idempotentAction("fail2ban_unban", func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	})

	body := `{"ip":"192.0.2.10","jail":"sshd"}`
	invalidMethod := idempotencyRequest(body, "req-00000003", "automation")
	invalidMethod.Method = http.MethodGet
	invalid := httptest.NewRecorder()
	handler(invalid, invalidMethod)
	if invalid.Code != http.StatusMethodNotAllowed {
		t.Fatalf("invalid method status = %d, want 405; body=%s", invalid.Code, invalid.Body.String())
	}
	if got := invalid.Header().Get("Allow"); got != http.MethodPost {
		t.Fatalf("Allow = %q, want POST", got)
	}

	valid := httptest.NewRecorder()
	handler(valid, idempotencyRequest(body, "req-00000003", "automation"))
	if valid.Code != http.StatusNoContent {
		t.Fatalf("valid POST status = %d, want 204; body=%s", valid.Code, valid.Body.String())
	}
	if calls != 1 {
		t.Fatalf("handler calls = %d, want 1", calls)
	}
}
