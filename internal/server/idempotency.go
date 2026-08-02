package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"regexp"

	"minion/internal/storage"
)

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{7,127}$`)

type bufferedResponseWriter struct {
	header      http.Header
	status      int
	wroteHeader bool
	body        bytes.Buffer
	parent      http.ResponseWriter
}

func newBufferedResponseWriter(parent http.ResponseWriter) *bufferedResponseWriter {
	return &bufferedResponseWriter{header: make(http.Header), status: http.StatusOK, parent: parent}
}

func (w *bufferedResponseWriter) Header() http.Header { return w.header }

func (w *bufferedResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
}

func (w *bufferedResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.body.Write(data)
}

func (w *bufferedResponseWriter) setClientName(name string) {
	setClientNameOnWriter(w.parent, name)
}

func (w *bufferedResponseWriter) setAuditDetail(action, target, detail string) {
	setAuditDetailOnWriter(w.parent, action, target, detail)
}

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func payloadDigest(body []byte) string {
	sum := sha256.Sum256(bytes.TrimSpace(body))
	return hex.EncodeToString(sum[:])
}

func replayBody(status int, body []byte) []byte {
	if status >= http.StatusBadRequest {
		return []byte("{\"error\":\"request failed\"}\n")
	}
	return append([]byte(nil), body...)
}

func (s *Server) idempotentAction(action string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		requestID := r.Header.Get("X-Request-ID")
		if !requestIDPattern.MatchString(requestID) {
			s.writeError(w, http.StatusBadRequest, "valid X-Request-ID header required")
			return
		}
		w.Header().Set("X-Request-ID", requestID)

		clientName := clientNameFromContext(r)
		if clientName == "" || s.storage == nil {
			s.writeError(w, http.StatusServiceUnavailable, "idempotency backend unavailable")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "failed to read request body")
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		record, err := s.storage.ClaimIdempotency(clientName, action, requestID, payloadDigest(body))
		if errors.Is(err, storage.ErrIdempotencyPayloadMismatch) {
			setAuditDetailOnWriter(w, action+"_idempotency_conflict", "", "request_id="+requestID)
			s.writeError(w, http.StatusConflict, "request id already used with a different payload")
			return
		}
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "failed to claim request id")
			return
		}

		switch record.State {
		case storage.IdempotencyCompleted:
			setAuditDetailOnWriter(w, action+"_replay", "", "request_id="+requestID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(record.StatusCode)
			_, _ = w.Write(record.ResponseBody)
			return
		case storage.IdempotencyInProgress:
			setAuditDetailOnWriter(w, action+"_in_progress", "", "request_id="+requestID)
			s.writeError(w, http.StatusConflict, "request with this id is already in progress")
			return
		case storage.IdempotencyClaimed:
			// Continue below.
		default:
			s.writeError(w, http.StatusInternalServerError, "invalid idempotency state")
			return
		}

		buffered := newBufferedResponseWriter(w)
		next(buffered, r)
		responseBody := replayBody(buffered.status, buffered.body.Bytes())
		if err := s.storage.CompleteIdempotency(clientName, action, requestID, buffered.status, responseBody); err != nil {
			s.writeError(w, http.StatusInternalServerError, "failed to persist request result")
			return
		}

		copyHeaders(w.Header(), buffered.Header())
		w.Header().Set("X-Request-ID", requestID)
		w.WriteHeader(buffered.status)
		_, _ = w.Write(buffered.body.Bytes())
	}
}
