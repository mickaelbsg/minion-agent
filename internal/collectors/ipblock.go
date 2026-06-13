package collectors

import (
    "bytes"
    "os/exec"
    "strings"
)

// IsIPBlocked checks whether the given IP appears in iptables rules or in /etc/hosts.deny.
// Returns true if the IP is found in either source, false otherwise.
func IsIPBlocked(ip string) (bool, error) {
    // 1. Check iptables rules (requires sudo for full list; we assume sudoers without password is configured).
    // Run with sudo to guarantee access.
    cmd := exec.Command("sudo", "iptables", "-nL")
    out, err := cmd.CombinedOutput()
    if err == nil {
        if bytes.Contains(out, []byte(ip)) {
            return true, nil
        }
    } else {
        // If sudo fails (e.g., not required), fall back to plain iptables.
        cmd = exec.Command("iptables", "-nL")
        out, err = cmd.CombinedOutput()
        if err == nil && bytes.Contains(out, []byte(ip)) {
            return true, nil
        }
    }

    // 2. Check hosts.deny file.
    data, err := exec.Command("cat", "/etc/hosts.deny").CombinedOutput()
    if err == nil {
        // Split lines, ignore comments and empty lines.
        lines := strings.Split(string(data), "\n")
        for _, line := range lines {
            trimmed := strings.TrimSpace(line)
            if trimmed == "" || strings.HasPrefix(trimmed, "#") {
                continue
            }
            // hosts.deny format: daemon : client [ : client ] ...
            // We'll simply search for the IP token.
            if strings.Contains(trimmed, ip) {
                return true, nil
            }
        }
    }
    // Not found in any source.
    return false, nil
}
