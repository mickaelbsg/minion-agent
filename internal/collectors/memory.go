package collectors

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type MemoryInfo struct {
	Total     uint64 `json:"total"`
	Free      uint64 `json:"free"`
	Available uint64 `json:"available"`
	Used      uint64 `json:"used"`
}

func GetMemory() (*MemoryInfo, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	return parseMemory(file)
}

func parseMemory(reader io.Reader) (*MemoryInfo, error) {
	info := &MemoryInfo{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSuffix(parts[0], ":")
		if key != "MemTotal" && key != "MemFree" && key != "MemAvailable" {
			continue
		}

		value, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid %s value %q: %w", key, parts[1], err)
		}

		switch key {
		case "MemTotal":
			info.Total = value
		case "MemFree":
			info.Free = value
		case "MemAvailable":
			info.Available = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read memory information: %w", err)
	}
	if info.Available > info.Total {
		return nil, fmt.Errorf("invalid memory information: available memory exceeds total memory")
	}

	info.Used = info.Total - info.Available
	return info, nil
}
