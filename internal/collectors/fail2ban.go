package collectors

import (
	"strings"
	"time"
)

type Fail2BanEvent struct {
	IP        string `json:"ip"`
	Jail      string `json:"jail"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
}

// GetFail2BanEvents returns the list of currently banned IPs per jail.
func GetFail2BanEvents() ([]Fail2BanEvent, error) {
	var result []Fail2BanEvent

	// Get list of jails
	out, err := runCommandOutput("fail2ban-client", "status")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	var jails []string
	for _, l := range lines {
		if strings.Contains(l, "Jail list:") {
			parts := strings.SplitN(l, ":", 2)
			if len(parts) == 2 {
				jails = strings.Fields(strings.TrimSpace(parts[1]))
			}
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	for _, jail := range jails {
		jailOut, err := runCommandOutput("fail2ban-client", "status", jail)
		if err != nil {
			continue // skip jail on error
		}
		jLines := strings.Split(string(jailOut), "\n")
		for _, jl := range jLines {
			if strings.Contains(jl, "Banned IP list:") {
				parts := strings.SplitN(jl, ":", 2)
				if len(parts) == 2 {
					ips := strings.Fields(strings.TrimSpace(parts[1]))
					for _, ip := range ips {
						result = append(result, Fail2BanEvent{IP: ip, Jail: jail, Action: "ban", Timestamp: now})
					}
				}
			}
		}
	}
	return result, nil
}

func UnbanFail2BanIP(jail, ip string) ([]byte, error) {
	return runCommandCombinedOutput("fail2ban-client", "set", jail, "unbanip", ip)
}
