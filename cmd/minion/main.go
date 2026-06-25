package main

import (
	"flag"
	"fmt"
	"log"
	"minion/internal/config"
	"minion/internal/security"
	"minion/internal/server"
	"minion/internal/storage"
	"os"
	"os/exec"
	"strings"
)

func main() {
	fs := flag.NewFlagSet("minion", flag.ExitOnError)
	configPath := fs.String("config", "/etc/minion/config.json", "Path to JSON configuration file")
	createClient := fs.Bool("create-client", false, "Create a new API client and print the API key")
	clientName := fs.String("name", "", "Name of the client")
	clientIPs := fs.String("ips", "", "Comma separated list of allowed IPs/CIDRs")

	// Lógica para permitir flags em qualquer lugar:
	// Coletamos todos os argumentos e separamos subcomandos de flags
	var subcommands []string
	var flagsOnly []string
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "-") {
			flagsOnly = append(flagsOnly, arg)
			// Se a flag tem um valor (não é booleana), pegamos o próximo também
			if !strings.Contains(arg, "=") && i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "-") {
				flagsOnly = append(flagsOnly, os.Args[i+1])
				i++
			}
		} else {
			subcommands = append(subcommands, arg)
		}
	}

	// Parseamos apenas as flags coletadas
	_ = fs.Parse(flagsOnly)

	// Se houver subcomandos
	if len(subcommands) > 0 {
		switch subcommands[0] {
		case "setup":
			setup(*configPath, *clientName, *clientIPs)
			return
		case "client":
			cmdArgs := []string{}
			if len(subcommands) > 1 {
				cmdArgs = subcommands[1:]
			}
			handleClientCommands(cmdArgs, *configPath, *clientName, *clientIPs)
			return
		case "add":
			if len(subcommands) > 1 && subcommands[1] == "client" {
				handleClientCommands([]string{"create"}, *configPath, *clientName, *clientIPs)
				return
			}
		}
	}

	if *createClient {
		handleClientCommands([]string{"create"}, *configPath, *clientName, *clientIPs)
		return
	}

	// Default: Start server
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	stor, err := storage.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to initialise storage: %v", err)
	}

	srv := server.New(cfg, stor)
	if err := srv.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func setup(configPath, clientName, clientIPs string) {
	if os.Geteuid() != 0 {
		log.Fatal("setup must be run as root. Use: sudo minion setup")
	}

	if err := os.MkdirAll("/etc/minion/tls", 0755); err != nil {
		log.Fatalf("failed to create /etc/minion/tls: %v", err)
	}
	certPath := "/etc/minion/tls/minion.crt"
	keyPath := "/etc/minion/tls/minion.key"
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		cmd := exec.Command("openssl", "req", "-newkey", "rsa:2048", "-nodes",
			"-keyout", keyPath, "-x509", "-days", "365", "-out", certPath,
			"-subj", "/CN=minion")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("failed to generate TLS cert: %v", err)
		}
		log.Printf("Generated self-signed TLS cert at %s", certPath)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if err := os.Chmod(configPath, 0600); err != nil {
		log.Printf("warning: failed to set config permissions: %v", err)
	}

	stor, err := storage.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to initialise storage: %v", err)
	}
	if err := os.Chmod(cfg.DBPath, 0600); err != nil {
		log.Printf("warning: failed to set database permissions: %v", err)
	}

	clients, err := stor.GetClients()
	if err != nil {
		log.Fatalf("failed to list clients: %v", err)
	}

	var generatedKey string
	createdClient := ""
	createdIPs := ""
	if len(clients) == 0 {
		createdClient = clientName
		if createdClient == "" {
			createdClient = "default"
		}
		createdIPs = clientIPs
		if createdIPs == "" {
			createdIPs = "127.0.0.1/32"
		}

		key, err := security.GenerateAPIKey()
		if err != nil {
			log.Fatalf("failed to generate API key: %v", err)
		}
		generatedKey = key
		hash := security.HashAPIKey(key)
		if err := stor.InsertClient(createdClient, createdIPs, hash); err != nil {
			log.Fatalf("failed to create bootstrap client: %v", err)
		}
	}

	if err := exec.Command("systemctl", "enable", "--now", "minion.service").Run(); err != nil {
		log.Printf("warning: failed to enable/start systemd service: %v", err)
	} else {
		log.Printf("systemd minion.service enabled and started")
	}

	fmt.Println("\nMinion setup completed.")
	if generatedKey != "" {
		fmt.Printf("Bootstrap client: %s\n", createdClient)
		fmt.Printf("Allowed IPs: %s\n", createdIPs)
		fmt.Printf("API Key: %s\n", generatedKey)
		fmt.Println("\nStore this API key now. It is shown only once and only its hash is stored.")
	} else {
		fmt.Println("Existing API clients found. No new bootstrap API key was generated.")
	}
}

func handleClientCommands(args []string, configPath, name, ips string) {
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	} else {
		cmd = "create"
	}

	cfg, _ := config.Load(configPath)
	dbPath := "/etc/minion/minion.db"
	if cfg != nil && cfg.DBPath != "" {
		dbPath = cfg.DBPath
	}
	stor, err := storage.New(dbPath)
	if err != nil {
		log.Fatalf("failed to open storage: %v", err)
	}

	switch cmd {
	case "create":
		if name == "" || ips == "" {
			log.Fatal("--name and --ips are required. Example: minion add client --name severino --ips 127.0.0.1/32")
		}
		key, err := security.GenerateAPIKey()
		if err != nil {
			log.Fatalf("failed to generate API key: %v", err)
		}
		hash := security.HashAPIKey(key)
		if err := stor.InsertClient(name, ips, hash); err != nil {
			log.Fatalf("failed to create client: %v", err)
		}
		fmt.Printf("Client: %s\nAPI Key: %s\nAPI Key Hash: %s\n", name, key, hash)
	case "list":
		clients, _ := stor.GetClients()
		fmt.Printf("%-20s %-30s %-10s\n", "NAME", "ALLOWED IPS", "ENABLED")
		for _, c := range clients {
			fmt.Printf("%-20s %-30s %-10v\n", c.Name, strings.Join(c.AllowedIPs, ","), c.Enabled)
		}
	case "enable":
		if len(args) < 2 {
			log.Fatal("client name required")
		}
		stor.UpdateClientStatus(args[1], true)
		fmt.Printf("Client %s enabled\n", args[1])
	case "disable":
		if len(args) < 2 {
			log.Fatal("client name required")
		}
		stor.UpdateClientStatus(args[1], false)
		fmt.Printf("Client %s disabled\n", args[1])
	case "delete":
		if len(args) < 2 {
			log.Fatal("client name required")
		}
		stor.DeleteClient(args[1])
		fmt.Printf("Client %s deleted\n", args[1])
	default:
		fmt.Println("Usage: minion client [create|list|enable|disable|delete]")
	}
}
