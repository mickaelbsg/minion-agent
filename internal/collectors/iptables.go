package collectors

import (
	"os/exec"
	"strings"
)

type IPTablesRule struct {
	Chain  string `json:"chain"`
	Target string `json:"target"`
	Source string `json:"source"`
	Dest   string `json:"destination"`
	Extra  string `json:"extra"`
}

func GetIPTablesRules() ([]IPTablesRule, error) {
	// Requer privilégios de root, mas como o agente rodará como systemd/root, deve funcionar.
	cmd := exec.Command("sudo", "iptables", "-S")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var rules []IPTablesRule
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		rule := IPTablesRule{
			Chain: "UNKNOWN",
		}
		
		// Parsing simplificado de iptables -S
		for i := 0; i < len(parts); i++ {
			switch parts[i] {
			case "-A", "-N", "-P":
				rule.Chain = parts[i+1]
			case "-j":
				rule.Target = parts[i+1]
			case "-s":
				rule.Source = parts[i+1]
			case "-d":
				rule.Dest = parts[i+1]
			}
		}
		rules = append(rules, rule)
	}

	return rules, nil
}
