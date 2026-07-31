package agentinfo

import (
	"sort"
	"strings"
	"testing"
)

func TestDeriveAgentIDIsStableAndDoesNotExposeSeed(t *testing.T) {
	seed := "0123456789abcdef0123456789abcdef"

	first := deriveAgentID(seed)
	second := deriveAgentID(seed)

	if first != second {
		t.Fatalf("expected stable agent id, got %q and %q", first, second)
	}
	if !strings.HasPrefix(first, "minion_") {
		t.Fatalf("expected minion_ prefix, got %q", first)
	}
	if strings.Contains(first, seed) {
		t.Fatal("agent id must not expose the machine id seed")
	}
}

func TestDeriveAgentIDUsesSafeFallbackForEmptySeed(t *testing.T) {
	if got := deriveAgentID(""); got == "" || got == "minion_" {
		t.Fatalf("expected non-empty fallback id, got %q", got)
	}
}

func TestCapabilitiesAreSortedAndContainCoreObservability(t *testing.T) {
	got := capabilities()
	if !sort.StringsAreSorted(got) {
		t.Fatalf("capabilities must be sorted: %v", got)
	}

	required := map[string]bool{
		"agent.read":              false,
		"users.read":              false,
		"journal.read":            false,
		"firewall.iptables.read": false,
		"fail2ban.read":           false,
		"fail2ban.unban":          false,
	}
	for _, capability := range got {
		if _, ok := required[capability]; ok {
			required[capability] = true
		}
	}
	for capability, found := range required {
		if !found {
			t.Fatalf("missing required capability %q", capability)
		}
	}
}
