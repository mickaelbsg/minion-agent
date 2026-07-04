package main

import (
	"flag"
	"fmt"
	"log"
	"minion/internal/admin"
	"minion/internal/config"
	"minion/internal/server"
	"minion/internal/storage"
	"minion/internal/ui"
	"os"
	"strings"
)

func main() {
	fs := flag.NewFlagSet("minion", flag.ExitOnError)
	configPath := fs.String("config", "/etc/minion/config.json", "Path to JSON configuration file")
	createClient := fs.Bool("create-client", false, "Create a new API client and print the API key")
	clientName := fs.String("name", "", "Name of the client")
	clientIPs := fs.String("ips", "", "Comma separated list of allowed IPs/CIDRs")
	uiSection := fs.String("section", "", "UI section: setup, config, clients, status")

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
		case "ui":
			if err := ui.Run(*configPath, *uiSection); err != nil {
				log.Fatalf("failed to start UI: %v", err)
			}
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
	service := admin.NewService(configPath)
	result, err := service.Setup(admin.SetupOptions{
		ClientName: clientName,
		ClientIPs:  clientIPs,
	})
	if err != nil {
		log.Fatalf("setup failed: %v", err)
	}

	fmt.Println("\nMinion setup completed.")
	if result.BootstrapCreated {
		fmt.Printf("Bootstrap client: %s\n", result.ClientName)
		fmt.Printf("Allowed IPs: %s\n", result.ClientIPs)
		fmt.Printf("API Key: %s\n", result.APIKey)
		fmt.Println("\nStore this API key now. It is shown only once and only its hash is stored.")
	} else {
		fmt.Println("Existing API clients found. No new bootstrap API key was generated.")
	}
}

func handleClientCommands(args []string, configPath, name, ips string) {
	service := admin.NewService(configPath)
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	} else {
		cmd = "create"
	}

	switch cmd {
	case "create":
		if name == "" || ips == "" {
			log.Fatal("--name and --ips are required. Example: minion add client --name severino --ips 127.0.0.1/32")
		}
		client, err := service.CreateClient(name, ips)
		if err != nil {
			log.Fatalf("failed to create client: %v", err)
		}
		fmt.Printf("Client: %s\nAPI Key: %s\nAPI Key Hash: %s\n", client.Name, client.APIKey, client.APIKeyHash)
	case "list":
		clients, err := service.ListClients()
		if err != nil {
			log.Fatalf("failed to list clients: %v", err)
		}
		fmt.Printf("%-20s %-30s %-10s\n", "NAME", "ALLOWED IPS", "ENABLED")
		for _, c := range clients {
			fmt.Printf("%-20s %-30s %-10v\n", c.Name, strings.Join(c.AllowedIPs, ","), c.Enabled)
		}
	case "enable":
		if len(args) < 2 {
			log.Fatal("client name required")
		}
		if err := service.SetClientEnabled(args[1], true); err != nil {
			log.Fatalf("failed to enable client: %v", err)
		}
		fmt.Printf("Client %s enabled\n", args[1])
	case "disable":
		if len(args) < 2 {
			log.Fatal("client name required")
		}
		if err := service.SetClientEnabled(args[1], false); err != nil {
			log.Fatalf("failed to disable client: %v", err)
		}
		fmt.Printf("Client %s disabled\n", args[1])
	case "delete":
		if len(args) < 2 {
			log.Fatal("client name required")
		}
		if err := service.DeleteClient(args[1]); err != nil {
			log.Fatalf("failed to delete client: %v", err)
		}
		fmt.Printf("Client %s deleted\n", args[1])
	default:
		fmt.Println("Usage: minion client [create|list|enable|disable|delete]")
	}
}
