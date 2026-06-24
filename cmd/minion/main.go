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
			setup()
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

func setup() {
	if err := os.MkdirAll("/etc/minion/tls", 0755); err != nil {
		log.Fatalf("failed to create /etc/minion/tls: %v", err)
	}
	certPath := "/etc/minion/tls/minion.crt"
	keyPath := "/etc/minion/tls/minion.key"
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		cmd := exec.Command("sudo", "openssl", "req", "-newkey", "rsa:2048", "-nodes",
			"-keyout", keyPath, "-x509", "-days", "365", "-out", certPath,
			"-subj", "/CN=minion")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("failed to generate TLS cert: %v", err)
		}
		log.Printf("Generated self‑signed TLS cert at %s", certPath)
	}
	if _, err := os.Stat("/etc/minion/config.json"); os.IsNotExist(err) {
		src := "config.example.json"
		if _, err := os.Stat(src); err == nil {
			data, _ := os.ReadFile(src)
			if err := os.WriteFile("/etc/minion/config.json", data, 0644); err != nil {
				log.Fatalf("failed to write config.json: %v", err)
			}
			log.Printf("Wrote default config to /etc/minion/config.json")
		}
	}
	if err := exec.Command("sudo", "systemctl", "enable", "--now", "minion.service").Run(); err != nil {
		log.Printf("warning: failed to enable/start systemd service: %v", err)
	} else {
		log.Printf("systemd minion.service enabled and started")
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
		key, _ := security.GenerateAPIKey()
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
