package main

import (
	"errors"
	"fmt"
	"log"
	"minion/internal/admin"
	"minion/internal/bootstrap"
	"os"
	"strings"
)

func dispatchCommand(subcommands []string, configPath, clientName, clientIPs, uiSection string) bool {
	if len(subcommands) == 0 {
		return false
	}
	switch subcommands[0] {
	case "setup":
		setup(configPath, clientName, clientIPs)
	case "bootstrap":
		handleBootstrapCommands(subcommands[1:], configPath, clientIPs)
	case "ui":
		runUI(configPath, uiSection)
	case "client":
		handleClientCommands(subcommands[1:], configPath, clientName, clientIPs)
	case "add":
		if len(subcommands) > 1 && subcommands[1] == "client" {
			handleClientCommands([]string{"create"}, configPath, clientName, clientIPs)
		} else {
			return false
		}
	default:
		return false
	}
	return true
}

func handleBootstrapCommands(args []string, configPath, ips string) {
	if len(args) != 1 {
		fmt.Println("Usage: sudo minion bootstrap show | sudo minion bootstrap pair --ips <ip/cidr>")
		return
	}

	switch args[0] {
	case "show":
		err := bootstrap.Consume(bootstrap.DefaultCredentialsPath, os.Stdout, isRoot)
		if errors.Is(err, bootstrap.ErrAlreadyConsumed) {
			log.Fatal("bootstrap credentials are unavailable or were already consumed; create a new client with `sudo minion client create --name <name> --ips <ip/cidr>`")
		}
		if err != nil {
			log.Fatalf("failed to show bootstrap credentials: %v", err)
		}
		fmt.Println("Bootstrap credentials consumed and removed from disk. Store the API key in the Automation credential store now.")
	case "pair":
		if strings.TrimSpace(ips) == "" {
			log.Fatal("--ips is required. Example: sudo minion bootstrap pair --ips 192.0.2.10/32")
		}
		payload, err := bootstrap.Read(bootstrap.DefaultCredentialsPath, isRoot)
		if errors.Is(err, bootstrap.ErrAlreadyConsumed) {
			log.Fatal("bootstrap credentials are unavailable or were already consumed; create a new client with `sudo minion client create --name automation --ips <ip/cidr>`")
		}
		if err != nil {
			log.Fatalf("failed to read bootstrap credentials: %v", err)
		}
		if err := admin.NewService(configPath).PairBootstrap(ips); err != nil {
			log.Fatalf("failed to pair bootstrap client: %v", err)
		}
		if _, err := os.Stdout.Write(payload); err != nil {
			log.Fatalf("bootstrap client was paired, but credentials could not be displayed; the credential file was preserved: %v", err)
		}
		if err := bootstrap.Remove(bootstrap.DefaultCredentialsPath); err != nil {
			log.Fatalf("bootstrap client was paired and credentials displayed, but the credential file could not be removed: %v", err)
		}
		fmt.Printf("Bootstrap paired for %s. Store the API key in the Automation credential store now.\n", ips)
	default:
		fmt.Println("Usage: sudo minion bootstrap show | sudo minion bootstrap pair --ips <ip/cidr>")
	}
}

func isRoot() bool { return os.Geteuid() == 0 }

func setup(configPath, clientName, clientIPs string) {
	service := admin.NewService(configPath)
	result, err := service.Setup(admin.SetupOptions{ClientName: clientName, ClientIPs: clientIPs})
	if err != nil {
		log.Fatalf("setup failed: %v", err)
	}
	fmt.Println("\nMinion setup completed.")
	if result.BootstrapCreated {
		err := bootstrap.WriteCredentials(
			bootstrap.DefaultCredentialsPath,
			result.ClientName,
			result.ClientIPs,
			result.APIKey,
		)
		if err != nil {
			if bootstrap.WasPublished(err) {
				log.Fatalf("setup created the client and published its credential, but could not confirm directory durability; the client was preserved for recovery: %v", err)
			}
			rollbackErr := service.DeleteClient(result.ClientName)
			if rollbackErr != nil {
				log.Fatalf("setup could not store bootstrap credentials and could not roll back bootstrap client: %v; rollback: %v", err, rollbackErr)
			}
			log.Fatalf("setup could not store bootstrap credentials; bootstrap client was rolled back: %v", err)
		}
		fmt.Printf("Bootstrap client: %s\nAllowed IPs: %s\n", result.ClientName, result.ClientIPs)
		fmt.Printf("Bootstrap credential stored root-only at %s.\n", bootstrap.DefaultCredentialsPath)
		if result.ClientName == "bootstrap" {
			fmt.Println("Use `sudo minion bootstrap pair --ips <AUTOMATION_IP/32>` to display it once and authorize Automation.")
		} else {
			fmt.Println("Use `sudo minion bootstrap show` to display it once. This client already uses the IP/CIDR configured during setup.")
		}
	} else {
		fmt.Println("Existing API clients found. No new bootstrap API key was generated.")
	}
}

func handleClientCommands(args []string, configPath, name, ips string) {
	service := admin.NewService(configPath)
	cmd := "create"
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "create":
		if name == "" || ips == "" {
			log.Fatal("--name and --ips are required. Example: minion client create --name severino --ips 127.0.0.1/32")
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
		fmt.Printf("%-20s %-30s %-10s %-25s %-25s\n", "NAME", "ALLOWED IPS", "STATUS", "EXPIRES AT", "REVOKED AT")
		for _, c := range clients {
			expiresAt := "never"
			if c.ExpiresAt != nil {
				expiresAt = c.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00")
			}
			status := "disabled"
			if c.Enabled {
				status = "enabled"
			}
			revokedAt := "-"
			if c.RevokedAt != nil {
				status = "revoked"
				revokedAt = c.RevokedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
			}
			fmt.Printf("%-20s %-30s %-10s %-25s %-25s\n", c.Name, strings.Join(c.AllowedIPs, ","), status, expiresAt, revokedAt)
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
	case "rotate":
		if len(args) < 2 {
			log.Fatal("client name required. Example: sudo minion client rotate automation")
		}
		apiKey, err := service.RotateClientAPIKey(args[1])
		if err != nil {
			log.Fatalf("failed to rotate client API key: %v", err)
		}
		fmt.Printf("Client: %s\nNew API Key: %s\n", args[1], apiKey)
		fmt.Println("The previous API key is now invalid. Update the credential in Automation/n8n immediately; this key will not be shown again.")
	case "revoke":
		if len(args) < 2 {
			log.Fatal("client name required. Example: sudo minion client revoke automation")
		}
		if err := service.RevokeClient(args[1]); err != nil {
			log.Fatalf("failed to revoke client: %v", err)
		}
		fmt.Printf("Client %s permanently revoked. Its API key is invalid and the record was preserved for audit.\n", args[1])
	case "expire":
		if len(args) < 3 {
			log.Fatal("client name and expiration required. Example: sudo minion client expire automation 2026-08-31T23:59:59Z; use never to remove expiration")
		}
		if err := service.SetClientExpiration(args[1], args[2]); err != nil {
			log.Fatalf("failed to set client expiration: %v", err)
		}
		if strings.EqualFold(args[2], "never") {
			fmt.Printf("Client %s no longer expires\n", args[1])
		} else {
			fmt.Printf("Client %s expires at %s\n", args[1], args[2])
		}
	case "delete":
		if len(args) < 2 {
			log.Fatal("client name required")
		}
		if err := service.DeleteClient(args[1]); err != nil {
			log.Fatalf("failed to delete client: %v", err)
		}
		fmt.Printf("Client %s deleted\n", args[1])
	default:
		fmt.Println("Usage: minion client [create|list|enable|disable|rotate|revoke|expire|delete]")
	}
}
