package server

import (
	"bytes"
	"io"
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
			s.writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBodyBytes+1))
		if err != nil {
			s.writeError(w, http.StatusBadRequest, "failed to read request body")
			return
		}
		_ = r.Body.Close()

		if len(body) > maxRequestBodyBytes {
			s.writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		next.ServeHTTP(w, r)
	})
}
