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

		// Tenta autenticar via Config estática primeiro
		authenticated := false
		for _, c := range s.cfg.Clients {
			if c.Enabled && security.IPAllowed(host, c.AllowedIPs) && security.VerifyAPIKey(apiKey, c.APIKeyHash) {
				r = withClientName(r, c.Name)
				authenticated = true
				break
			}
		}

		// Se não autenticou via config, tenta via Banco de Dados
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

// ... (restantes handlers handleHealth, handleSystem, etc permanecem os mesmos)
