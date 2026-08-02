package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestListInProgressIdempotencyFiltersOrdersAndPaginates(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "minion.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.DB.Close()

	for _, tc := range []struct {
		client, action, requestID string
	}{
		{"automation-b", "fail2ban_unban", "req-new"},
		{"automation-a", "fail2ban_unban", "req-old"},
		{"automation-a", "other_action", "req-other"},
		{"automation-a", "fail2ban_unban", "req-completed"},
	} {
		if _, err := store.ClaimIdempotency(tc.client, tc.action, tc.requestID, "secret-payload-hash"); err != nil {
			t.Fatalf("claim %s error = %v", tc.requestID, err)
		}
	}
	if err := store.CompleteIdempotency("automation-a", "fail2ban_unban", "req-completed", 200, []byte(`{"secret":"response"}`)); err != nil {
		t.Fatalf("complete error = %v", err)
	}

	old := time.Now().UTC().Add(-2 * time.Hour)
	if _, err := store.DB.Exec(`UPDATE idempotency_records SET created_at = ?, updated_at = ? WHERE request_id = ?`, old, old, "req-old"); err != nil {
		t.Fatalf("age record error = %v", err)
	}

	first, err := store.ListInProgressIdempotency("fail2ban_unban", 1, 0)
	if err != nil {
		t.Fatalf("first page error = %v", err)
	}
	second, err := store.ListInProgressIdempotency("fail2ban_unban", 1, 1)
	if err != nil {
		t.Fatalf("second page error = %v", err)
	}
	if len(first) != 1 || first[0].RequestID != "req-old" {
		t.Fatalf("unexpected first page: %+v", first)
	}
	if len(second) != 1 || second[0].RequestID != "req-new" {
		t.Fatalf("unexpected second page: %+v", second)
	}
}

func TestListInProgressIdempotencyValidatesBounds(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "minion.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.DB.Close()

	for _, tc := range []struct{ limit, offset int }{{0, 0}, {maxIdempotencyDiagnosticsQueryLimit + 1, 0}, {1, -1}} {
		if _, err := store.ListInProgressIdempotency("", tc.limit, tc.offset); err == nil {
			t.Fatalf("expected error for limit=%d offset=%d", tc.limit, tc.offset)
		}
	}
}
