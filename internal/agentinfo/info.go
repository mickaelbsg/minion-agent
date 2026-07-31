package agentinfo

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
)

const unknownValue = "unknown"

// Info describes the local Minion instance and the capabilities exposed by it.
type Info struct {
	AgentID     string   `json:"agent_id"`
	Hostname    string   `json:"hostname"`
	Version     string   `json:"version"`
	OS          string   `json:"os"`
	Architecture string   `json:"architecture"`
	UptimeSeconds float64 `json:"uptime_seconds"`
	ObservedAt  string   `json:"observed_at"`
	Capabilities []string `json:"capabilities"`
}

// Get returns metadata that a control plane can use to identify and inspect this Minion.
func Get() Info {
	hostname, _ := os.Hostname()
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		hostname = unknownValue
	}

	seed := readMachineID()
	if seed == "" {
		seed = hostname
	}

	return Info{
		AgentID:       deriveAgentID(seed),
		Hostname:      hostname,
		Version:       buildVersion(),
		OS:            runtime.GOOS,
		Architecture:  runtime.GOARCH,
		UptimeSeconds: readUptime(),
		ObservedAt:    time.Now().UTC().Format(time.RFC3339),
		Capabilities:  capabilities(),
	}
}

func deriveAgentID(seed string) string {
	seed = strings.TrimSpace(seed)
	if seed == "" {
		seed = unknownValue
	}
	sum := sha256.Sum256([]byte("minion-agent:" + seed))
	return "minion_" + hex.EncodeToString(sum[:16])
}

func readMachineID() string {
	for _, path := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id"} {
		data, err := os.ReadFile(path)
		if err == nil {
			if value := strings.TrimSpace(string(data)); value != "" {
				return value
			}
		}
	}
	return ""
}

func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "devel"
	}
	return info.Main.Version
}

func readUptime() float64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0
	}
	value, err := strconv.ParseFloat(fields[0], 64)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func capabilities() []string {
	items := []string{
		"agent.read",
		"disk.read",
		"fail2ban.read",
		"fail2ban.unban",
		"firewall.iptables.read",
		"ipblock.read",
		"journal.read",
		"logins.read",
		"memory.read",
		"privilege-events.read",
		"services.read",
		"system.read",
		"users.read",
		"wazuh.read",
	}
	sort.Strings(items)
	return items
}
