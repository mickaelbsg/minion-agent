package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestListInProgressIdempotencyFiltersOrdersAndSanitizes(t *testing.T) {
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

	items, err := store.ListInProgressIdempotency("fail2ban_unban", 10)
	if err != nil {
		t.Fatalf("ListInProgressIdempotency() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].RequestID != "req-old" || items[1].RequestID != "req-new" {
		t.Fatalf("unexpected order: %+v", items)
	}
	if items[0].ClientName != "automation-a" || items[0].Action != "fail2ban_unban" {
		t.Fatalf("unexpected diagnostic: %+v", items[0])
	}
}

func TestListInProgressIdempotencyEnforcesLimit(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "minion.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.DB.Close()

	for _, limit := range []int{0, MaxIdempotencyDiagnosticsLimit + 1} {
		if _, err := store.ListInProgressIdempotency("", limit); err == nil {
			t.Fatalf("expected error for limit %d", limit)
		}
	}
}
