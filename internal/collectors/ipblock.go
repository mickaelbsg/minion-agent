package collectors

import (
	"bytes"
	"os"
	"strings"
)

// IsIPBlocked checks whether the given IP appears in iptables rules or in /etc/hosts.deny.
// Returns true if the IP is found in either source, false otherwise.
func IsIPBlocked(ip string) (bool, error) {
	// The service is expected to run with the required system privileges via systemd.
	// Do not call sudo from inside the agent runtime.
	out, err := runCommandCombinedOutput("iptables", "-nL")
	if err == nil && bytes.Contains(out, []byte(ip)) {
		return true, nil
	}

	data, readErr := os.ReadFile("/etc/hosts.deny")
	if readErr == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			if strings.Contains(trimmed, ip) {
				return true, nil
			}
		}
	}

	// Keep the collector tolerant: inability to read iptables or hosts.deny means
	// the IP was not confirmed as blocked by this lightweight check.
	return false, nil
}
