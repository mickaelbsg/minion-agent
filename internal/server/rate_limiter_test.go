package server

import (
	"testing"
	"time"
)

func TestRateLimiterBlocksAndRecovers(t *testing.T) {
	limiter := newRateLimiter(2, 1)
	now := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }
	limiter.lastCleanup = now

	if allowed, _ := limiter.allow("client-a"); !allowed {
		t.Fatal("first request should be allowed")
	}
	if allowed, _ := limiter.allow("client-a"); !allowed {
		t.Fatal("second request should be allowed")
	}
	if allowed, retry := limiter.allow("client-a"); allowed || retry < time.Second {
		t.Fatalf("third request should be limited with retry >= 1s, allowed=%v retry=%s", allowed, retry)
	}

	now = now.Add(time.Second)
	if allowed, _ := limiter.allow("client-a"); !allowed {
		t.Fatal("request should recover after one token is refilled")
	}
}

func TestRateLimiterIsolatesKeys(t *testing.T) {
	limiter := newRateLimiter(1, 1)
	now := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }
	limiter.lastCleanup = now

	if allowed, _ := limiter.allow("client-a"); !allowed {
		t.Fatal("client-a first request should be allowed")
	}
	if allowed, _ := limiter.allow("client-a"); allowed {
		t.Fatal("client-a second request should be limited")
	}
	if allowed, _ := limiter.allow("client-b"); !allowed {
		t.Fatal("client-b should have an independent bucket")
	}
}

func TestRateLimiterCleansInactiveEntries(t *testing.T) {
	limiter := newRateLimiter(1, 1)
	now := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }
	limiter.lastCleanup = now

	limiter.allow("stale")
	now = now.Add(7 * time.Minute)
	limiter.allow("active")

	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	if _, exists := limiter.entries["stale"]; exists {
		t.Fatal("inactive entry should be removed during cleanup")
	}
}
