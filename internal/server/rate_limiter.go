package server

import (
	"math"
	"sync"
	"time"
)

type rateLimitEntry struct {
	tokens   float64
	lastSeen time.Time
}

type rateLimiter struct {
	mu          sync.Mutex
	capacity    float64
	refillRate  float64
	entries     map[string]rateLimitEntry
	now         func() time.Time
	lastCleanup time.Time
}

func newRateLimiter(capacity int, refillPerSecond float64) *rateLimiter {
	if capacity < 1 {
		capacity = 1
	}
	if refillPerSecond <= 0 {
		refillPerSecond = 1
	}
	now := time.Now
	return &rateLimiter{
		capacity:    float64(capacity),
		refillRate:  refillPerSecond,
		entries:     make(map[string]rateLimitEntry),
		now:         now,
		lastCleanup: now(),
	}
}

func (l *rateLimiter) allow(key string) (bool, time.Duration) {
	now := l.now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastCleanup) >= 5*time.Minute {
		l.cleanup(now)
	}

	entry, ok := l.entries[key]
	if !ok {
		entry = rateLimitEntry{tokens: l.capacity, lastSeen: now}
	} else {
		elapsed := now.Sub(entry.lastSeen).Seconds()
		entry.tokens = math.Min(l.capacity, entry.tokens+elapsed*l.refillRate)
		entry.lastSeen = now
	}

	if entry.tokens >= 1 {
		entry.tokens--
		entry.lastSeen = now
		l.entries[key] = entry
		return true, 0
	}

	entry.lastSeen = now
	l.entries[key] = entry
	missing := 1 - entry.tokens
	retry := time.Duration(math.Ceil(missing/l.refillRate*1000)) * time.Millisecond
	if retry < time.Second {
		retry = time.Second
	}
	return false, retry
}

func (l *rateLimiter) cleanup(now time.Time) {
	idleTTL := time.Duration(math.Ceil(l.capacity/l.refillRate))*time.Second + 5*time.Minute
	for key, entry := range l.entries {
		if now.Sub(entry.lastSeen) > idleTTL {
			delete(l.entries, key)
		}
	}
	l.lastCleanup = now
}
