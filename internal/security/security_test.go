package security

import "testing"

func TestHashAndVerifyAPIKey(t *testing.T) {
    key := "my-secret-key"
    hash := HashAPIKey(key)
    if !VerifyAPIKey(key, hash) {
        t.Fatalf("expected VerifyAPIKey to succeed for matching key")
    }
    if VerifyAPIKey("wrong-key", hash) {
        t.Fatalf("expected VerifyAPIKey to fail for mismatched key")
    }
}

func TestIPAllowed(t *testing.T) {
    allow := []string{"192.168.1.0/24", "10.0.0.5"}
    tests := []struct {
        ip     string
        expect bool
    }{
        {"192.168.1.42", true},
        {"10.0.0.5", true},
        {"10.0.0.6", false},
        {"invalid-ip", false},
    }
    for _, tt := range tests {
        if got := IPAllowed(tt.ip, allow); got != tt.expect {
            t.Fatalf("IPAllowed(%s) = %v, want %v", tt.ip, got, tt.expect)
        }
    }
}
