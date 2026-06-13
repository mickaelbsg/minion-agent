package server

import (
    "context"
    "net"
    "net/http"
)

type contextKey string

const clientNameKey contextKey = "clientName"

// withClientName stores the authenticated client name in the request context.
func withClientName(r *http.Request, name string) *http.Request {
    ctx := context.WithValue(r.Context(), clientNameKey, name)
    return r.WithContext(ctx)
}

// clientNameFromContext retrieves the client name from the request context, if any.
func clientNameFromContext(r *http.Request) string {
    if v := r.Context().Value(clientNameKey); v != nil {
        if s, ok := v.(string); ok {
            return s
        }
    }
    return ""
}

type statusRecorder struct {
    http.ResponseWriter
    status int
}

func (rec *statusRecorder) WriteHeader(code int) {
    rec.status = code
    rec.ResponseWriter.WriteHeader(code)
}

// audit is a middleware that records each request to the storage layer.
func (s *Server) audit(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
        // Call the next handler.
        next(rec, r)
        // Gather audit data.
        clientName := clientNameFromContext(r)
        host, _, err := net.SplitHostPort(r.RemoteAddr)
        if err != nil {
            host = r.RemoteAddr
        }
        // Insert the audit record; ignore error as logging is best‑effort.
        _ = s.storage.InsertAudit(clientName, host, r.Method, r.URL.Path, rec.status)
    }
}
