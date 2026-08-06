package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"minion/internal/admin"
	"minion/internal/config"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const packageClientAbsentExitCode = 3

const (
	packageTLSCertPath = "/etc/minion/tls/minion.crt"
	packageTLSKeyPath  = "/etc/minion/tls/minion.key"
)

func handlePackageCommands(args []string, configPath, clientName string) {
	if !isRoot() {
		log.Fatal("package commands require root privileges")
	}
	if len(args) != 1 {
		log.Fatal("usage: sudo minion package <client-exists|ensure-tls|ready> [options]")
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
	case "ensure-tls":
		created, err := packageEnsureTLS(packageTLSCertPath, packageTLSKeyPath, time.Now())
		if err != nil {
			log.Fatalf("failed to ensure package TLS assets: %v", err)
		}
		if created {
			fmt.Println("created")
		} else {
			fmt.Println("preserved")
		}
	case "ready":
		if err := packageReady(configPath); err != nil {
			log.Fatalf("package readiness check failed: %v", err)
		}
		fmt.Println("ready")
	default:
		log.Fatal("usage: sudo minion package <client-exists|ensure-tls|ready> [options]")
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

func packageEnsureTLS(certPath, keyPath string, now time.Time) (bool, error) {
	certExists, err := packageSafeRegularFileExists(certPath)
	if err != nil {
		return false, err
	}
	keyExists, err := packageSafeRegularFileExists(keyPath)
	if err != nil {
		return false, err
	}
	if certExists != keyExists {
		return false, fmt.Errorf("incomplete TLS pair: certificate and private key must both exist or both be absent")
	}
	if certExists {
		if _, err := tls.LoadX509KeyPair(certPath, keyPath); err != nil {
			return false, fmt.Errorf("existing TLS pair is invalid: %w", err)
		}
		return false, nil
	}

	dir := filepath.Dir(certPath)
	if filepath.Dir(keyPath) != dir {
		return false, fmt.Errorf("certificate and private key must share the same directory")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return false, fmt.Errorf("create TLS directory: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return false, fmt.Errorf("secure TLS directory: %w", err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return false, fmt.Errorf("generate private key: %w", err)
	}
	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return false, fmt.Errorf("generate certificate serial: %w", err)
	}
	template := x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "minion"},
		NotBefore:    now.Add(-5 * time.Minute),
		NotAfter:     now.Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"minion", "localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return false, fmt.Errorf("create certificate: %w", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
		return false, fmt.Errorf("validate generated TLS pair: %w", err)
	}

	keyTemp, err := packageWriteTempFile(dir, ".minion.key-*", keyPEM, 0o600)
	if err != nil {
		return false, err
	}
	defer os.Remove(keyTemp)
	certTemp, err := packageWriteTempFile(dir, ".minion.crt-*", certPEM, 0o644)
	if err != nil {
		return false, err
	}
	defer os.Remove(certTemp)

	if err := os.Rename(keyTemp, keyPath); err != nil {
		return false, fmt.Errorf("publish private key: %w", err)
	}
	if err := os.Rename(certTemp, certPath); err != nil {
		_ = os.Remove(keyPath)
		return false, fmt.Errorf("publish certificate: %w", err)
	}
	if _, err := tls.LoadX509KeyPair(certPath, keyPath); err != nil {
		_ = os.Remove(certPath)
		_ = os.Remove(keyPath)
		return false, fmt.Errorf("validate published TLS pair: %w", err)
	}
	return true, nil
}

func packageSafeRegularFileExists(path string) (bool, error) {
	info, err := os.Lstat(path)
	if errorsIsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("refusing unsafe TLS path %s", path)
	}
	return true, nil
}

func errorsIsNotExist(err error) bool {
	return err != nil && os.IsNotExist(err)
}

func packageWriteTempFile(dir, pattern string, content []byte, mode os.FileMode) (path string, resultErr error) {
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", fmt.Errorf("create temporary TLS file: %w", err)
	}
	path = file.Name()
	defer func() {
		if closeErr := file.Close(); resultErr == nil && closeErr != nil {
			resultErr = fmt.Errorf("close temporary TLS file: %w", closeErr)
		}
		if resultErr != nil {
			_ = os.Remove(path)
		}
	}()
	if err := file.Chmod(mode); err != nil {
		return "", fmt.Errorf("set temporary TLS permissions: %w", err)
	}
	if _, err := file.Write(content); err != nil {
		return "", fmt.Errorf("write temporary TLS file: %w", err)
	}
	if err := file.Sync(); err != nil {
		return "", fmt.Errorf("sync temporary TLS file: %w", err)
	}
	return path, nil
}

func packageReady(configPath string) error {
	cfg, err := config.Read(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	endpoint, err := packageReadinessEndpoint(cfg, packageTLSAssetsExist())
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

func packageTLSAssetsExist() bool {
	return packageRegularFileExists(packageTLSCertPath) && packageRegularFileExists(packageTLSKeyPath)
}

func packageRegularFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func packageReadinessEndpoint(cfg *config.Config, tlsAssetsExist bool) (string, error) {
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
	if cfg.API.AllowInsecureHTTP && !tlsAssetsExist {
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
