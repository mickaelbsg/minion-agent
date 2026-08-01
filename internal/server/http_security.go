package server

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"time"
)

const maxRequestBodyBytes = 64 * 1024
const readHeaderTimeout = 5 * time.Second
const readTimeout = 15 * time.Second
const writeTimeout = 30 * time.Second
const idleTimeout = 60 * time.Second

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

func (s *Server) limitRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil || r.Body == http.NoBody {
			next.ServeHTTP(w, r)
			return
		}

		if r.ContentLength > maxRequestBodyBytes {
			s.auditBodyRejection(r, http.StatusRequestEntityTooLarge, "declared_size_exceeded")
			s.writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBodyBytes+1))
		if err != nil {
			s.auditBodyRejection(r, http.StatusBadRequest, "body_read_failed")
			s.writeError(w, http.StatusBadRequest, "failed to read request body")
			return
		}
		_ = r.Body.Close()

		if len(body) > maxRequestBodyBytes {
			s.auditBodyRejection(r, http.StatusRequestEntityTooLarge, "streamed_size_exceeded")
			s.writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		next.ServeHTTP(w, r)
	})
}

func (s *Server) auditBodyRejection(r *http.Request, status int, reason string) {
	if s == nil || s.storage == nil {
		return
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	_ = s.storage.InsertAuditDetail(
		"",
		host,
		r.Method,
		r.URL.Path,
		status,
		"request_body_rejected",
		"",
		"reason="+reason,
	)
}
