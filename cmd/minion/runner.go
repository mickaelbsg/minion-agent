package main

import (
	"flag"
	"log"
	"minion/internal/config"
	"minion/internal/server"
	"minion/internal/storage"
	"os"
	"strings"
)

func run() {
	fs := flag.NewFlagSet("minion", flag.ExitOnError)
	configPath := fs.String("config", "/etc/minion/config.json", "Path to JSON configuration file")
	createClient := fs.Bool("create-client", false, "Create a new API client and print the API key")
	clientName := fs.String("name", "", "Name of the client")
	clientIPs := fs.String("ips", "", "Comma separated list of allowed IPs/CIDRs")
	uiSection := fs.String("section", "", "UI section: setup, config, clients, status")

	var subcommands []string
	var flagsOnly []string
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "-") {
			flagsOnly = append(flagsOnly, arg)
			if !strings.Contains(arg, "=") && i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "-") {
				flagsOnly = append(flagsOnly, os.Args[i+1])
				i++
			}
		} else {
			subcommands = append(subcommands, arg)
		}
	}
	_ = fs.Parse(flagsOnly)

	if len(subcommands) > 0 && subcommands[0] == "package" {
		handlePackageCommands(subcommands[1:], *configPath, *clientName)
		return
	}
	if dispatchCommand(subcommands, *configPath, *clientName, *clientIPs, *uiSection) {
		return
	}
	if *createClient {
		handleClientCommands([]string{"create"}, *configPath, *clientName, *clientIPs)
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
	if err = srv.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
