package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net"
	"strings"

	"golang.org/x/crypto/argon2"
)

func GenerateAPIKey() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "minion_sk_" + base64.RawURLEncoding.EncodeToString(buf), nil
}

// HashAPIKey generates a salted Argon2id hash for the given API key.
// A new random 16-byte salt is generated for each call and the result is
// returned in the form "base64(salt)$base64(hash)". This format allows the
// corresponding VerifyAPIKey function to recreate the hash using the stored
// salt.
func HashAPIKey(apiKey string) string {
	// Generate a random 16-byte salt.
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		// In the unlikely event of a RNG failure, fall back to a deterministic
		// salt so the function still returns something usable (the caller will
		// treat this as an internal error).
		salt = []byte("fallback_salt_123")
	}
	hash := argon2.IDKey([]byte(apiKey), salt, 1, 64*1024, 4, 32)
	// Encode both components using URL-safe base64 without padding.
	return base64.RawURLEncoding.EncodeToString(salt) + "$" + base64.RawURLEncoding.EncodeToString(hash)
}

// VerifyAPIKey checks whether the supplied apiKey matches the stored salted
// hash (in the format produced by HashAPIKey). It returns true on a match.
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
