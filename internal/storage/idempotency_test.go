package storage

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestIdempotencyClaimCompleteAndReplay(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "minion.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.DB.Close()

	first, err := store.ClaimIdempotency("automation", "fail2ban_unban", "req-0001", "hash-a")
	if err != nil {
		t.Fatalf("first claim error = %v", err)
	}
	if first.State != IdempotencyClaimed {
		t.Fatalf("first state = %q, want %q", first.State, IdempotencyClaimed)
	}

	pending, err := store.ClaimIdempotency("automation", "fail2ban_unban", "req-0001", "hash-a")
	if err != nil {
		t.Fatalf("pending claim error = %v", err)
	}
	if pending.State != IdempotencyInProgress {
		t.Fatalf("pending state = %q, want %q", pending.State, IdempotencyInProgress)
	}

	body := []byte(`{"status":"unbanned"}`)
	if err := store.CompleteIdempotency("automation", "fail2ban_unban", "req-0001", 200, body); err != nil {
		t.Fatalf("complete error = %v", err)
	}

	replay, err := store.ClaimIdempotency("automation", "fail2ban_unban", "req-0001", "hash-a")
	if err != nil {
		t.Fatalf("replay claim error = %v", err)
	}
	if replay.State != IdempotencyCompleted || replay.StatusCode != 200 || string(replay.ResponseBody) != string(body) {
		t.Fatalf("unexpected replay record: %+v", replay)
	}
}

func TestIdempotencyRejectsPayloadMismatchAndIsolatesClients(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "minion.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.DB.Close()

	if _, err := store.ClaimIdempotency("automation-a", "fail2ban_unban", "req-0002", "hash-a"); err != nil {
		t.Fatalf("claim error = %v", err)
	}
	if _, err := store.ClaimIdempotency("automation-a", "fail2ban_unban", "req-0002", "hash-b"); !errors.Is(err, ErrIdempotencyPayloadMismatch) {
		t.Fatalf("mismatch error = %v, want ErrIdempotencyPayloadMismatch", err)
	}

	other, err := store.ClaimIdempotency("automation-b", "fail2ban_unban", "req-0002", "hash-b")
	if err != nil {
		t.Fatalf("other client claim error = %v", err)
	}
	if other.State != IdempotencyClaimed {
		t.Fatalf("other client state = %q, want %q", other.State, IdempotencyClaimed)
	}
}

func TestPurgeCompletedIdempotencyBeforePreservesRecentAndInProgress(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "minion.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.DB.Close()

	for _, requestID := range []string{"completed-old", "completed-recent", "in-progress-old"} {
		if _, err := store.ClaimIdempotency("automation", "fail2ban_unban", requestID, "hash-"+requestID); err != nil {
			t.Fatalf("claim %s error = %v", requestID, err)
		}
	}
	if err := store.CompleteIdempotency("automation", "fail2ban_unban", "completed-old", 200, []byte(`{"status":"ok"}`)); err != nil {
		t.Fatalf("complete old error = %v", err)
	}
	if err := store.CompleteIdempotency("automation", "fail2ban_unban", "completed-recent", 200, []byte(`{"status":"ok"}`)); err != nil {
		t.Fatalf("complete recent error = %v", err)
	}

	old := time.Now().UTC().Add(-8 * 24 * time.Hour)
	if _, err := store.DB.Exec(
		`UPDATE idempotency_records SET updated_at = ? WHERE request_id IN (?, ?)`,
		old, "completed-old", "in-progress-old",
	); err != nil {
		t.Fatalf("age records error = %v", err)
	}

	removed, err := store.PurgeCompletedIdempotencyBefore(time.Now().UTC().Add(-7 * 24 * time.Hour))
	if err != nil {
		t.Fatalf("purge error = %v", err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}

	var oldCompletedCount, recentCompletedCount, oldInProgressCount int
	if err := store.DB.QueryRow(`SELECT COUNT(*) FROM idempotency_records WHERE request_id = ?`, "completed-old").Scan(&oldCompletedCount); err != nil {
		t.Fatalf("count old completed error = %v", err)
	}
	if err := store.DB.QueryRow(`SELECT COUNT(*) FROM idempotency_records WHERE request_id = ?`, "completed-recent").Scan(&recentCompletedCount); err != nil {
		t.Fatalf("count recent completed error = %v", err)
	}
	if err := store.DB.QueryRow(`SELECT COUNT(*) FROM idempotency_records WHERE request_id = ?`, "in-progress-old").Scan(&oldInProgressCount); err != nil {
		t.Fatalf("count old in-progress error = %v", err)
	}

	if oldCompletedCount != 0 || recentCompletedCount != 1 || oldInProgressCount != 1 {
		t.Fatalf("unexpected retention result: old completed=%d recent completed=%d old in-progress=%d", oldCompletedCount, recentCompletedCount, oldInProgressCount)
	}
}

func TestPurgeCompletedIdempotencyBeforeRejectsZeroCutoff(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "minion.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.DB.Close()

	if _, err := store.PurgeCompletedIdempotencyBefore(time.Time{}); err == nil {
		t.Fatal("expected zero cutoff error")
	}
}
