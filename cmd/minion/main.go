package main

import (
    "flag"
    "log"

    "minion/internal/config"
    "minion/internal/security"
    "minion/internal/server"
    "minion/internal/storage"
    "os"
    "os/exec"
)

func main() {
    configPath := flag.String("config", "/etc/minion/config.json", "Path to JSON configuration file")
    createClient := flag.Bool("create-client", false, "Create a new API client and print the API key")
    clientName := flag.String("name", "", "Name of the client to create")
    clientIPs := flag.String("ips", "", "Comma separated list of allowed IPs/CIDRs")
    flag.Parse()

// Handle subcommands after flag parsing
args := flag.Args()
if len(args) > 0 {
    switch args[0] {
    case "setup":
        // Perform initial setup: create directories, generate TLS cert, copy example config, enable service
        if err := os.MkdirAll("/etc/minion/tls", 0755); err != nil {
            log.Fatalf("failed to create /etc/minion/tls: %v", err)
        }
        // Generate TLS cert if missing
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
        } else {
            log.Printf("TLS cert already exists at %s", certPath)
        }
        // Copy example config if config not present
        if _, err := os.Stat("/etc/minion/config.json"); os.IsNotExist(err) {
            // Try to copy from example path relative to binary location
            src := "config.example.json"
            if _, err := os.Stat(src); err == nil {
                data, _ := os.ReadFile(src)
                if err := os.WriteFile("/etc/minion/config.json", data, 0644); err != nil {
                    log.Fatalf("failed to write config.json: %v", err)
                }
                log.Printf("Wrote default config to /etc/minion/config.json")
            } else {
                log.Printf("Example config %s not found, skipping config copy", src)
            }
        }
        // Enable and start systemd service
        if err := exec.Command("sudo", "systemctl", "enable", "--now", "minion.service").Run(); err != nil {
            log.Printf("warning: failed to enable/start systemd service: %v", err)
        } else {
            log.Printf("systemd minion.service enabled and started")
        }
        return
    case "add":
        if len(args) > 1 && args[1] == "client" {
            // reuse existing create-client logic below
            if *clientName == "" || *clientIPs == "" {
                log.Fatalf("--name and --ips must be provided with add client")
            }
            key, err := security.GenerateAPIKey()
            if err != nil {
                log.Fatalf("failed to generate API key: %v", err)
            }
            hash := security.HashAPIKey(key)
            log.Printf("Client: %s", *clientName)
            log.Printf("Allowed IPs: %s", *clientIPs)
            log.Printf("API Key: %s", key)
            log.Printf("API Key Hash (to place in config): %s", hash)
            return
        }
    }
}


    if *createClient {
        if *clientName == "" || *clientIPs == "" {
            log.Fatalf("--name and --ips must be provided with --create-client")
        }

        key, err := security.GenerateAPIKey()
        if err != nil {
            log.Fatalf("failed to generate API key: %v", err)
        }

        hash := security.HashAPIKey(key)

        log.Printf("Client: %s", *clientName)
        log.Printf("Allowed IPs: %s", *clientIPs)
        log.Printf("API Key: %s", key)
        log.Printf("API Key Hash (to place in config): %s", hash)
        return
    }

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