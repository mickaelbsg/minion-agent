package collectors

import (
	"os/exec"
	"strconv"
	"strings"
)

type JournalEntry struct {
	Timestamp string `json:"timestamp"`
	Host      string `json:"host"`
	Process   string `json:"process"`
	Message   string `json:"message"`
}

func GetJournalLogs(limit string, level string) ([]JournalEntry, error) {
	args := []string{"--no-pager"}

	if limit == "" {
		limit = "100"
	}
	if n, err := strconv.Atoi(limit); err == nil && n > 0 && n <= 1000 {
		args = append(args, "-n", limit)
	} else {
		args = append(args, "-n", "100")
	}

	if level != "" {
		args = append(args, "-p", level)
	}

	// The service is expected to run with the required system privileges via systemd.
	// Do not call sudo from inside the agent runtime.
	cmd := exec.Command("journalctl", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return ParseJournalLogs(string(output)), nil
}

func ParseJournalLogs(output string) []JournalEntry {
	var entries []JournalEntry
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}

		message := strings.Join(parts[4:], " ")
		entry := JournalEntry{
			Timestamp: strings.Join(parts[0:3], " "),
			Host:      parts[3],
			Message:   message,
		}

		if idx := strings.Index(parts[4], ":"); idx != -1 {
			entry.Process = parts[4][:idx]
			if len(message) > idx+2 {
				entry.Message = message[idx+2:]
			}
		}

		entries = append(entries, entry)
	}
	return entries
}
