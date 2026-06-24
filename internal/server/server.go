package server

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os/exec"
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

	mux.HandleFunc("/api/v1/health", s.audit(s.handleHealth))
	mux.HandleFunc("/api/v1/system", s.audit(s.auth(s.handleSystem)))
	mux.HandleFunc("/api/v1/users", s.audit(s.auth(s.handleUsers)))
	mux.HandleFunc("/api/v1/services", s.audit(s.auth(s.handleServices)))
	mux.HandleFunc("/api/v1/fail2ban", s.audit(s.auth(s.handleFail2Ban)))
	mux.HandleFunc("/api/v1/fail2ban/unban", s.audit(s.auth(s.handleFail2BanUnban)))
	mux.HandleFunc("/api/v1/ipblock", s.audit(s.auth(s.handleIPBlock)))
	mux.HandleFunc("/api/v1/wazuh", s.audit(s.auth(s.handleWazuh)))
	mux.HandleFunc("/api/v1/logins", s.audit(s.auth(s.handleLogins)))
	mux.HandleFunc("/api/v1/memory", s.audit(s.auth(s.handleMemory)))
	mux.HandleFunc("/api/v1/iptables", s.audit(s.auth(s.handleIPTables)))

	addr := s.cfg.API.Bind
	if addr == "" {
		addr = "0.0.0.0:9870"
	}

	log.Printf("Minion listening on %s", addr)
	certFile := "/etc/minion/tls/minion.crt"
	keyFile := "/etc/minion/tls/minion.key"

	// Fallback para HTTP se os arquivos TLS não existirem (para facilitar testes)
	if _, err := net.LookupHost("localhost"); err == nil {
		return http.ListenAndServeTLS(addr, certFile, keyFile, mux)
	}
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

		authenticated := false
		for _, c := range s.cfg.Clients {
			if c.Enabled && security.IPAllowed(host, c.AllowedIPs) && security.VerifyAPIKey(apiKey, c.APIKeyHash) {
				r = withClientName(r, c.Name)
				authenticated = true
				break
			}
		}

		if !authenticated && s.storage != nil {
			clients, err := s.storage.GetClients()
			if err == nil {
				for _, c := range clients {
					if security.IPAllowed(host, c.AllowedIPs) && security.VerifyAPIKey(apiKey, c.APIKeyHash) {
						r = withClientName(r, c.Name)
						authenticated = true
						break
					}
				}
			}
		}

		if !authenticated {
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
	services, err := collectors.GetServices()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, services)
}

func (s *Server) handleFail2Ban(w http.ResponseWriter, r *http.Request) {
	items, err := collectors.GetFail2BanStatus()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, items)
}

func (s *Server) handleFail2BanUnban(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var payload struct {
		IP   string `json:"ip"`
		Jail string `json:"jail"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	if net.ParseIP(payload.IP) == nil {
		s.writeError(w, http.StatusBadRequest, "invalid ip address")
		return
	}
	allowedJails := []string{"sshd", "apache-auth", "recidive"}
	jailAllowed := false
	for _, aj := range allowedJails {
		if payload.Jail == aj {
			jailAllowed = true
			break
		}
	}
	if !jailAllowed {
		s.writeError(w, http.StatusBadRequest, "jail not allowed")
		return
	}
	cmd := exec.Command("fail2ban-client", "set", payload.Jail, "unbanip", payload.IP)
	out, err := cmd.CombinedOutput()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, string(out))
		return
	}
	s.writeJSON(w, map[string]string{"status": "unbanned", "ip": payload.IP, "jail": payload.Jail})
}

func (s *Server) handleIPBlock(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		s.writeError(w, http.StatusBadRequest, "ip query parameter required")
		return
	}
	blocked, err := collectors.IsIPBlocked(ip)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, map[string]bool{"blocked": blocked})
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

func (s *Server) handleMemory(w http.ResponseWriter, r *http.Request) {
	mem, err := collectors.GetMemory()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, mem)
}

func (s *Server) handleIPTables(w http.ResponseWriter, r *http.Request) {
	rules, err := collectors.GetIPTablesRules()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, rules)
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
