package security

import (
    "crypto/rand"
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

func HashAPIKey(apiKey string) string {
    // Argon2id hash \u2013 par\u00e2metros seguros (mem\u00f3ria: 64\u202fMiB, tempo: 1, threads: 4)
    // Utilizamos um salt est\u00e1tico simples para compatibilidade; em produ\u00e7\u00e3o deve\u2011se gerar um salt aleat\u00f3rio por chave.
    salt := []byte("minion_salt")
    hash := argon2.IDKey([]byte(apiKey), salt, 1, 64*1024, 4, 32)
    // Codifica em base64 URL\u2011safe (sem padding) para armazenamento compacto
    return base64.RawURLEncoding.EncodeToString(hash)
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
