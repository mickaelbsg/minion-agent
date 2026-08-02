package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"minion/internal/storage"
)

func newIdempotencyDiagnosticsServer(t *testing.T) *Server {
	t.Helper()
	store, err := storage.New(filepath.Join(t.TempDir(), "minion.db"))
	if err != nil {
		t.Fatalf("storage.New() error = %v", err)
	}
	t.Cleanup(func() { _ = store.DB.Close() })
	return &Server{storage: store}
}

func TestHandleIdempotencyInProgressReturnsSanitizedRecords(t *testing.T) {
	s := newIdempotencyDiagnosticsServer(t)
	if _, err := s.storage.ClaimIdempotency("automation", "fail2ban_unban", "req-00000010", "payload-secret"); err != nil {
		t.Fatalf("claim error = %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/idempotency/in-progress?action=fail2ban_unban&limit=10", nil)
	s.handleIdempotencyInProgress(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Count      int                             `json:"count"`
		Offset     int                             `json:"offset"`
		NextOffset int                             `json:"next_offset"`
		HasMore    bool                            `json:"has_more"`
		Records    []storage.IdempotencyDiagnostic `json:"records"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if response.Count != 1 || response.Offset != 0 || response.NextOffset != 1 || response.HasMore {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.Records[0].RequestID != "req-00000010" {
		t.Fatalf("request_id = %q", response.Records[0].RequestID)
	}
	body := recorder.Body.String()
	for _, forbidden := range []string{"payload-secret", "payload_hash", "response_body"} {
		if stringContains(body, forbidden) {
			t.Fatalf("response contains forbidden value %q: %s", forbidden, body)
		}
	}
}

func TestHandleIdempotencyInProgressPaginatesBeyondFirstPage(t *testing.T) {
	s := newIdempotencyDiagnosticsServer(t)
	for i := 0; i < 3; i++ {
		requestID := fmt.Sprintf("req-%02d", i)
		if _, err := s.storage.ClaimIdempotency("automation", "fail2ban_unban", requestID, "hash"); err != nil {
			t.Fatalf("claim error = %v", err)
		}
	}

	first := httptest.NewRecorder()
	s.handleIdempotencyInProgress(first, httptest.NewRequest(http.MethodGet, "/api/v1/idempotency/in-progress?limit=2", nil))
	var page struct {
		Count      int  `json:"count"`
		NextOffset int  `json:"next_offset"`
		HasMore    bool `json:"has_more"`
	}
	if err := json.Unmarshal(first.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode first page error = %v", err)
	}
	if page.Count != 2 || !page.HasMore || page.NextOffset != 2 {
		t.Fatalf("unexpected first page: %+v", page)
	}

	second := httptest.NewRecorder()
	s.handleIdempotencyInProgress(second, httptest.NewRequest(http.MethodGet, "/api/v1/idempotency/in-progress?limit=2&offset=2", nil))
	if err := json.Unmarshal(second.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode second page error = %v", err)
	}
	if page.Count != 1 || page.HasMore || page.NextOffset != 3 {
		t.Fatalf("unexpected second page: %+v", page)
	}
}

func TestHandleIdempotencyInProgressValidatesMethodAndQuery(t *testing.T) {
	s := newIdempotencyDiagnosticsServer(t)

	post := httptest.NewRecorder()
	s.handleIdempotencyInProgress(post, httptest.NewRequest(http.MethodPost, "/api/v1/idempotency/in-progress", nil))
	if post.Code != http.StatusMethodNotAllowed || post.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("POST status=%d Allow=%q", post.Code, post.Header().Get("Allow"))
	}

	for _, target := range []string{
		"/api/v1/idempotency/in-progress?limit=0",
		"/api/v1/idempotency/in-progress?limit=101",
		"/api/v1/idempotency/in-progress?limit=invalid",
		"/api/v1/idempotency/in-progress?offset=-1",
		"/api/v1/idempotency/in-progress?offset=invalid",
		"/api/v1/idempotency/in-progress?action=unknown",
	} {
		recorder := httptest.NewRecorder()
		s.handleIdempotencyInProgress(recorder, httptest.NewRequest(http.MethodGet, target, nil))
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("target %s status=%d, want 400", target, recorder.Code)
		}
	}
}

func stringContains(value, fragment string) bool {
	for i := 0; i+len(fragment) <= len(value); i++ {
		if value[i:i+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
