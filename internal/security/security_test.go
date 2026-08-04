package security

import (
	"errors"
	"testing"
)

func TestHashAndVerifyAPIKey(t *testing.T) {
	key := "my-secret-key"
	hash, err := HashAPIKeyWithError(key)
	if err != nil {
		t.Fatalf("HashAPIKeyWithError returned error: %v", err)
	}
	if !VerifyAPIKey(key, hash) {
		t.Fatalf("expected VerifyAPIKey to succeed for matching key")
	}
	if VerifyAPIKey("wrong-key", hash) {
		t.Fatalf("expected VerifyAPIKey to fail for mismatched key")
	}
}

func TestHashAPIKeyFailsClosedWithoutEntropy(t *testing.T) {
	original := readRandom
	readRandom = func([]byte) (int, error) {
		return 0, errors.New("entropy unavailable")
	}
	t.Cleanup(func() { readRandom = original })

	hash, err := HashAPIKeyWithError("secret")
	if err == nil {
		t.Fatal("expected entropy failure to be returned")
	}
	if hash != "" {
		t.Fatalf("expected no hash on entropy failure, got %q", hash)
	}
	if legacy := HashAPIKey("secret"); legacy != "" {
		t.Fatalf("expected compatibility wrapper to fail closed, got %q", legacy)
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
