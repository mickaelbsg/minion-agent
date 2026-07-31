package collectors

import (
	"strings"
	"testing"
)

func TestParseMemory(t *testing.T) {
	input := strings.NewReader(`MemTotal:       16000 kB
MemFree:         4000 kB
MemAvailable:   10000 kB
Buffers:          500 kB
`)

	info, err := parseMemory(input)
	if err != nil {
		t.Fatalf("parseMemory returned an unexpected error: %v", err)
	}

	if info.Total != 16000 {
		t.Fatalf("Total = %d, want 16000", info.Total)
	}
	if info.Free != 4000 {
		t.Fatalf("Free = %d, want 4000", info.Free)
	}
	if info.Available != 10000 {
		t.Fatalf("Available = %d, want 10000", info.Available)
	}
	if info.Used != 6000 {
		t.Fatalf("Used = %d, want 6000", info.Used)
	}
}

func TestParseMemoryRejectsInvalidNumber(t *testing.T) {
	_, err := parseMemory(strings.NewReader("MemTotal: invalid kB\n"))
	if err == nil {
		t.Fatal("parseMemory returned nil error for an invalid numeric value")
	}
}

func TestParseMemoryRejectsAvailableGreaterThanTotal(t *testing.T) {
	input := strings.NewReader(`MemTotal:       1000 kB
MemAvailable:   2000 kB
`)

	_, err := parseMemory(input)
	if err == nil {
		t.Fatal("parseMemory returned nil error when available memory exceeded total memory")
	}
}
