package collectors

import "strings"

type IPTablesRule struct {
	Chain  string `json:"chain"`
	Target string `json:"target"`
	Source string `json:"source"`
	Dest   string `json:"destination"`
	Extra  string `json:"extra"`
}

func GetIPTablesRules() ([]IPTablesRule, error) {
	// The agent is expected to run with the required system privileges via systemd.
	// Avoid calling sudo here because sudo can block or fail when no TTY/sudoers rule exists.
	output, err := runCommandOutput("iptables", "-S")
	if err != nil {
		return nil, err
	}

	return ParseIPTablesRules(string(output)), nil
}

func ParseIPTablesRules(output string) []IPTablesRule {
	var rules []IPTablesRule
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		rule := IPTablesRule{Chain: "UNKNOWN"}

		for i := 0; i < len(parts); i++ {
			switch parts[i] {
			case "-A", "-N", "-P":
				if i+1 < len(parts) {
					rule.Chain = parts[i+1]
				}
			case "-j":
				if i+1 < len(parts) {
					rule.Target = parts[i+1]
				}
			case "-s":
				if i+1 < len(parts) {
					rule.Source = parts[i+1]
				}
			case "-d":
				if i+1 < len(parts) {
					rule.Dest = parts[i+1]
				}
			}
		}

		rule.Extra = line
		rules = append(rules, rule)
	}

	return rules
}
