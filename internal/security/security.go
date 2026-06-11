package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net"
	"strings"
)

func GenerateAPIKey() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "minion_sk_" + base64.RawURLEncoding.EncodeToString(buf), nil
}

func HashAPIKey(apiKey string) string {
	sum := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(sum[:])
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
