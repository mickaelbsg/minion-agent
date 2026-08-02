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

	action := strings.TrimSpace(r.URL.Query().Get("action"))
	if action != "" && action != "fail2ban_unban" {
		s.writeError(w, http.StatusBadRequest, "unsupported action filter")
		return
	}

	items, err := s.storage.ListInProgressIdempotency(action, limit)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "failed to query idempotency diagnostics")
		return
	}

	s.writeJSON(w, map[string]interface{}{
		"count":   len(items),
		"records": items,
	})
}
