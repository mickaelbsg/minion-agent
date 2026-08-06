package tlsbootstrap

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// Ensure creates a secure self-signed bootstrap TLS pair when both files are absent.
// Existing valid pairs are preserved. Incomplete, invalid, symlinked or non-regular
// paths are rejected without overwriting operator-managed material.
func Ensure(certPath, keyPath string, now time.Time) (bool, error) {
	certExists, err := safeRegularFileExists(certPath)
	if err != nil {
		return false, err
	}
	keyExists, err := safeRegularFileExists(keyPath)
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

	keyTemp, err := writeTempFile(dir, ".minion.key-*", keyPEM, 0o600)
	if err != nil {
		return false, err
	}
	defer os.Remove(keyTemp)
	certTemp, err := writeTempFile(dir, ".minion.crt-*", certPEM, 0o644)
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

func safeRegularFileExists(path string) (bool, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
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

func writeTempFile(dir, pattern string, content []byte, mode os.FileMode) (path string, resultErr error) {
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
