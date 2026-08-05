package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"minion/internal/admin"
	"minion/internal/config"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const packageClientAbsentExitCode = 3

func handlePackageCommands(args []string, configPath, clientName string) {
	if !isRoot() {
		log.Fatal("package commands require root privileges")
	}
	if len(args) != 1 {
		log.Fatal("usage: sudo minion package <client-exists|ready> [options]")
	}

	switch args[0] {
	case "client-exists":
		if strings.TrimSpace(clientName) == "" {
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
	case "ready":
		if err := packageReady(configPath); err != nil {
			log.Fatalf("package readiness check failed: %v", err)
		}
		fmt.Println("ready")
	default:
		log.Fatal("usage: sudo minion package <client-exists|ready> [options]")
	}
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

func packageReady(configPath string) error {
	cfg, err := config.Read(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	endpoint, err := packageReadinessEndpoint(cfg)
	if err != nil {
		return err
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Local bootstrap certificate is self-signed.
	}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}
	return packageReadyWithClient(client, endpoint)
}

func packageReadinessEndpoint(cfg *config.Config) (string, error) {
	host, port, err := net.SplitHostPort(cfg.API.Bind)
	if err != nil {
		return "", fmt.Errorf("invalid api.bind %q: %w", cfg.API.Bind, err)
	}
	switch host {
	case "", "0.0.0.0":
		host = "127.0.0.1"
	case "::":
		host = "::1"
	}
	scheme := "https"
	if cfg.API.AllowInsecureHTTP {
		scheme = "http"
	}
	return scheme + "://" + net.JoinHostPort(host, port) + "/api/v1/health", nil
}

func packageReadyWithClient(client *http.Client, endpoint string) error {
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build health request: %w", err)
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("request health endpoint: %w", err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1024))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("health endpoint returned %s", response.Status)
	}
	return nil
}
