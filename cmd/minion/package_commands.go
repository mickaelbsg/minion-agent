package main

import (
	"fmt"
	"log"
	"minion/internal/admin"
	"os"
	"strings"
)

const packageClientAbsentExitCode = 3

func handlePackageCommands(args []string, configPath, clientName string) {
	if !isRoot() {
		log.Fatal("package commands require root privileges")
	}
	if len(args) != 1 || args[0] != "client-exists" || strings.TrimSpace(clientName) == "" {
		log.Fatal("usage: sudo minion package client-exists --name <client>")
	}

	exists, err := packageClientExists(admin.NewService(configPath), clientName)
	if err != nil {
		log.Fatalf("failed to inspect package client state: %v", err)
	}
	if !exists {
		os.Exit(packageClientAbsentExitCode)
	}
	fmt.Println("present")
}

func packageClientExists(service *admin.Service, name string) (bool, error) {
	clients, err := service.ListClients()
	if err != nil {
		return false, err
	}
	for _, client := range clients {
		if client.Name == name {
			return true, nil
		}
	}
	return false, nil
}
