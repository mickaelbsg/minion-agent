package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net"
	"strings"

	"golang.org/x/crypto/argon2"
)

var readRandom = rand.Read

func GenerateAPIKey() (string, error) {
	buf := make([]byte, 32)
	if _, err := readRandom(buf); err != nil {
		return "", err
	}
	return "minion_sk_" + base64.RawURLEncoding.EncodeToString(buf), nil
}

// HashAPIKeyWithError generates a salted Argon2id hash for the given API key.
// A new random 16-byte salt is generated for each call and the result is
// returned in the form "base64(salt)$base64(hash)". Entropy failures are
// returned to the caller so credentials are never created with a predictable
// fallback salt.
func HashAPIKeyWithError(apiKey string) (string, error) {
	salt := make([]byte, 16)
	if _, err := readRandom(salt); err != nil {
		return "", fmt.Errorf("generate Argon2id salt: %w", err)
	}
	hash := argon2.IDKey([]byte(apiKey), salt, 1, 64*1024, 4, 32)
	return base64.RawURLEncoding.EncodeToString(salt) + "$" + base64.RawURLEncoding.EncodeToString(hash), nil
}

// HashAPIKey is retained for compatibility with read-only tooling and tests.
// New credential-management code must use HashAPIKeyWithError and propagate
// failures. An empty string indicates that secure entropy was unavailable.
func HashAPIKey(apiKey string) string {
	hash, err := HashAPIKeyWithError(apiKey)
	if err != nil {
		return ""
	}
	return hash
}

// VerifyAPIKey checks whether the supplied apiKey matches the stored salted
// hash (in the format produced by HashAPIKeyWithError). It returns true on a match.
func VerifyAPIKey(apiKey, stored string) bool {
	parts := strings.SplitN(stored, "$", 2)
	if len(parts) != 2 {
		return false
	}
	saltBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	expectedHash, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	hash := argon2.IDKey([]byte(apiKey), saltBytes, 1, 64*1024, 4, 32)
	return subtle.ConstantTimeCompare(hash, expectedHash) == 1
}

func IPAllowed(ip string, allowList []string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	for _, entry := range allowList {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			_, network, err := net.ParseCIDR(entry)
			if err != nil {
				continue
			}
			if network.Contains(parsedIP) {
				return true
			}
			continue
		}
		if entry == parsedIP.String() {
			return true
		}
	}
	return false
}
