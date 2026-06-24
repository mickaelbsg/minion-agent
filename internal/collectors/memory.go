package collectors

import (
	"bufio"
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
	defer file.Close()

	info := &MemoryInfo{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSuffix(parts[0], ":")
		value, _ := strconv.ParseUint(parts[1], 10, 64)

		switch key {
		case "MemTotal":
			info.Total = value
		case "MemFree":
			info.Free = value
		case "MemAvailable":
			info.Available = value
		}
	}

	info.Used = info.Total - info.Available
	return info, nil
}
