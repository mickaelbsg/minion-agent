package collectors

import (
	"os/exec"
	"strings"
)

type SudoEvent struct {
	Timestamp string `json:"timestamp"`
	User      string `json:"user"`
	TTY       string `json:"tty"`
	PWD       string `json:"pwd"`
	Command   string `json:"command"`
}

func GetSudoEvents() ([]SudoEvent, error) {
	// Busca eventos de sudo no journalctl
	cmd := exec.Command("sudo", "journalctl", "_COMM=sudo", "-n", "50", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var events []SudoEvent
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" || !strings.Contains(line, "COMMAND=") {
			continue
		}
		
		// Parsing simplificado do log do sudo
		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}

		event := SudoEvent{
			Timestamp: strings.Join(parts[0:3], " "),
		}

		// Extrair detalhes do comando
		if idx := strings.Index(line, " ; "); idx != -1 {
			details := line[idx+3:]
			detailParts := strings.Split(details, " ; ")
			for _, dp := range detailParts {
				kv := strings.SplitN(dp, "=", 2)
				if len(kv) == 2 {
					switch kv[0] {
					case "USER":
						event.User = kv[1]
					case "TTY":
						event.TTY = kv[1]
					case "PWD":
						event.PWD = kv[1]
					case "COMMAND":
						event.Command = kv[1]
					}
				}
			}
		}
		events = append(events, event)
	}
	return events, nil
}
