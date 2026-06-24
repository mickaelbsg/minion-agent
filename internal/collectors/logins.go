package collectors

import (
	"os/exec"
	"strings"
)

type LoginEvent struct {
	User      string `json:"user"`
	IP        string `json:"ip"`
	Success   bool   `json:"success"`
	Timestamp string `json:"timestamp"`
}

func GetLogins() ([]LoginEvent, error) {
	out, err := exec.Command("last", "-w", "-n", "50", "--time-format", "iso").Output()
	if err != nil {
		return nil, err
	}

	logins := []LoginEvent{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "wtmp begins") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		user := fields[0]
		if user == "reboot" || user == "shutdown" || user == "runlevel" {
			continue
		}

		timestampIndex := findISOTimestampIndex(fields)
		if timestampIndex == -1 {
			continue
		}

		ip := "local"
		if timestampIndex > 2 {
			ip = fields[2]
		}

		logins = append(logins, LoginEvent{
			User:      user,
			IP:        ip,
			Success:   true,
			Timestamp: fields[timestampIndex],
		})
	}

	return logins, nil
}

func findISOTimestampIndex(fields []string) int {
	for i, field := range fields {
		if len(field) >= len("2006-01-02T15:04:05") && field[4] == '-' && field[7] == '-' && field[10] == 'T' {
			return i
		}
	}
	return -1
}
