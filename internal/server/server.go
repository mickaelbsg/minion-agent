package server

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"

	"minion/internal/collectors"
	"minion/internal/config"
	"minion/internal/security"
	"minion/internal/storage"
)

type Server struct {
	cfg     *config.Config
	storage *storage.Storage
}

func New(cfg *config.Config, storage *storage.Storage) *Server {
	return &Server{cfg: cfg, storage: storage}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", s.handleHealth)
	mux.HandleFunc("/api/v1/system", s.auth(s.handleSystem))
	mux.HandleFunc("/api/v1/users", s.auth(s.handleUsers))
	mux.HandleFunc("/api/v1/services", s.auth(s.handleServices))
	mux.HandleFunc("/api/v1/fail2ban", s.auth(s.handleFail2Ban))
	mux.HandleFunc("/api/v1/wazuh", s.auth(s.handleWazuh))
	mux.HandleFunc("/api/v1/logins", s.auth(s.handleLogins))

	addr := s.cfg.API.Bind
	if addr == "" {
		addr = "0.0.0.0:9870"
	}

	log.Printf("Minion listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}

		apiKey := r.Header.Get("Authorization")
		if strings.HasPrefix(strings.ToLower(apiKey), "bearer ") {
			apiKey = strings.TrimSpace(apiKey[7:])
		}

		if apiKey == "" {
			s.writeError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		hashed := security.HashAPIKey(apiKey)
		var client *config.Client

		for i := range s.cfg.Clients {
			c := &s.cfg.Clients[i]
			if !c.Enabled {
				continue
			}
			if !security.IPAllowed(host, c.AllowedIPs) {
				continue
			}
			if c.APIKeyHash == hashed {
				client = c
				break
			}
		}

		if client == nil {
			s.writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		next(w, r)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleSystem(w http.ResponseWriter, r *http.Request) {
	sys, err := collectors.GetSystem()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, sys)
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	users, err := collectors.GetUsers()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, users)
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	items, err := collectors.GetServices()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, items)
}

func (s *Server) handleFail2Ban(w http.ResponseWriter, r *http.Request) {
	items, err := collectors.GetFail2BanEvents()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, items)
}

func (s *Server) handleWazuh(w http.ResponseWriter, r *http.Request) {
	status, err := collectors.GetWazuhStatus()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, status)
}

func (s *Server) handleLogins(w http.ResponseWriter, r *http.Request) {
	items, err := collectors.GetLogins()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, items)
}

func (s *Server) writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to write json: %v", err)
	}
}

func (s *Server) writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
