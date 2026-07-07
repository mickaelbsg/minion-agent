package server

import (
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

type clientNameSetter interface {
	setClientName(string)
}

func setClientNameOnWriter(w http.ResponseWriter, name string) {
	if setter, ok := w.(clientNameSetter); ok {
		setter.setClientName(name)
	}
}

type auditDetailSetter interface {
	setAuditDetail(action, target, detail string)
}

func setAuditDetailOnWriter(w http.ResponseWriter, action, target, detail string) {
	if setter, ok := w.(auditDetailSetter); ok {
		setter.setAuditDetail(action, target, detail)
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status     int
	clientName string
	action     string
	target     string
	detail     string
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *statusRecorder) setClientName(name string) {
	rec.clientName = name
}

func (rec *statusRecorder) setAuditDetail(action, target, detail string) {
	rec.action = action
	rec.target = target
	rec.detail = detail
}

// audit is a middleware that records each request to the storage layer.
func (s *Server) audit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		// Call the next handler.
		next(rec, r)
		// Gather audit data.
		clientName := rec.clientName
		if clientName == "" {
			clientName = clientNameFromContext(r)
		}
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		// Insert the audit record; ignore error as logging is best-effort.
		if s.storage != nil {
			_ = s.storage.InsertAuditDetail(clientName, host, r.Method, r.URL.Path, rec.status, rec.action, rec.target, rec.detail)
		}
	}
}
