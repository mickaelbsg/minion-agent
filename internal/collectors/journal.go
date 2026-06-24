package collectors

import (
	"os/exec"
	"strings"
)

type JournalEntry struct {
	Timestamp string `json:"timestamp"`
	Host      string `json:"host"`
	Process   string `json:"process"`
	Message   string `json:"message"`
}

func GetJournalLogs(limit string, level string) ([]JournalEntry, error) {
	args := []string{"journalctl", "--no-pager"}
	if limit != "" {
		args = append(args, "-n", limit)
	} else {
		args = append(args, "-n", "100")
	}
	if level != "" {
		args = append(args, "-p", level)
	}

	cmd := exec.Command("sudo", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var entries []JournalEntry
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}

		entry := JournalEntry{
			Timestamp: strings.Join(parts[0:3], " "),
			Host:      parts[3],
			Message:   strings.Join(parts[4:], " "),
		}
		
		if idx := strings.Index(parts[4], ":"); idx != -1 {
			entry.Process = parts[4][:idx]
			entry.Message = strings.Join(parts[4:], " ")[idx+2:]
		}

		entries = append(entries, entry)
	}
	return entries, nil
}
