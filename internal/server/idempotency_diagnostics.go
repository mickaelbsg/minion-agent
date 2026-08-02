package server

import (
	"net/http"
	"strconv"
	"strings"

	"minion/internal/storage"
)

const defaultIdempotencyDiagnosticsLimit = 50

func (s *Server) handleIdempotencyInProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if s == nil || s.storage == nil {
		s.writeError(w, http.StatusServiceUnavailable, "idempotency storage unavailable")
		return
	}

	limit := defaultIdempotencyDiagnosticsLimit
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > storage.MaxIdempotencyDiagnosticsLimit {
			s.writeError(w, http.StatusBadRequest, "limit must be between 1 and 100")
			return
		}
		limit = parsed
	}

	offset := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			s.writeError(w, http.StatusBadRequest, "offset must be zero or greater")
			return
		}
		offset = parsed
	}

	action := strings.TrimSpace(r.URL.Query().Get("action"))
	if action != "" && action != "fail2ban_unban" {
		s.writeError(w, http.StatusBadRequest, "unsupported action filter")
		return
	}

	items, err := s.storage.ListInProgressIdempotency(action, limit+1, offset)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "failed to query idempotency diagnostics")
		return
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	nextOffset := offset + len(items)

	s.writeJSON(w, map[string]interface{}{
		"count":       len(items),
		"offset":      offset,
		"next_offset": nextOffset,
		"has_more":    hasMore,
		"records":     items,
	})
}
