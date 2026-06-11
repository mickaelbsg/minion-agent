package main

import (
    "flag"
    "log"

    "minion/internal/config"
    "minion/internal/security"
    "minion/internal/server"
    "minion/internal/storage"
)

func main() {
    configPath := flag.String("config", "/etc/minion/config.json", "Path to JSON configuration file")
    createClient := flag.Bool("create-client", false, "Create a new API client and print the API key")
    clientName := flag.String("name", "", "Name of the client to create")
    clientIPs := flag.String("ips", "", "Comma separated list of allowed IPs/CIDRs")
    flag.Parse()

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
