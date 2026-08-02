package server

import (
	"log"
	"time"
)

func (s *Server) purgeExpiredIdempotency() {
	if s == nil || s.storage == nil || s.cfg == nil {
		return
	}

	retention := time.Duration(s.cfg.Security.IdempotencyRetentionHours) * time.Hour
	cutoff := time.Now().UTC().Add(-retention)
	removed, err := s.storage.PurgeCompletedIdempotencyBefore(cutoff)
	if err != nil {
		log.Printf("failed to purge expired idempotency records: %v", err)
		return
	}
	if removed > 0 {
		log.Printf("purged %d expired completed idempotency records", removed)
	}
}
