package storage

import (
	"errors"
	"path/filepath"
	"testing"
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
