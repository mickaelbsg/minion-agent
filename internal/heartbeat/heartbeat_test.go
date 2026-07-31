package heartbeat

import (
	"testing"
	"time"
)

func TestAtReturnsOnlineHeartbeat(t *testing.T) {
	original := processStartedAt
	processStartedAt = time.Date(2026, 7, 31, 20, 0, 0, 0, time.UTC)
	t.Cleanup(func() { processStartedAt = original })

	got := At(processStartedAt.Add(90 * time.Second))
	if got.Status != "online" {
		t.Fatalf("expected online status, got %q", got.Status)
	}
	if got.AgentID == "" || got.Hostname == "" || got.Version == "" {
		t.Fatalf("expected agent identity fields, got %+v", got)
	}
	if got.ProcessUptimeSeconds != 90 {
		t.Fatalf("expected 90 seconds of process uptime, got %v", got.ProcessUptimeSeconds)
	}
}

func TestAtNeverReturnsNegativeProcessUptime(t *testing.T) {
	original := processStartedAt
	processStartedAt = time.Date(2026, 7, 31, 20, 0, 0, 0, time.UTC)
	t.Cleanup(func() { processStartedAt = original })

	got := At(processStartedAt.Add(-time.Minute))
	if got.ProcessUptimeSeconds != 0 {
		t.Fatalf("expected zero process uptime, got %v", got.ProcessUptimeSeconds)
	}
	if got.ObservedAt != got.ProcessStartedAt {
		t.Fatalf("expected observation clamped to process start, got %q and %q", got.ObservedAt, got.ProcessStartedAt)
	}
}
