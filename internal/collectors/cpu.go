package collectors

import (
	"os"
	"strings"
)

// CPUInfo representa informações básicas da CPU
type CPUInfo struct {
	ModelName string `json:"model_name"`
	Cores     int    `json:"cores"`
}

// GetCPU coleta informações básicas da CPU do sistema Linux
func GetCPU() (*CPUInfo, error) {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return nil, err
	}

	info := &CPUInfo{}
	lines := strings.Split(string(data), "\n")
	cores := 0

	for _, line := range lines {
		if strings.HasPrefix(line, "model name") {
			if info.ModelName == "" {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					info.ModelName = strings.TrimSpace(parts[1])
				}
			}
		}
		if strings.HasPrefix(line, "processor") {
			cores++
		}
	}

	info.Cores = cores
	return info, nil
}
