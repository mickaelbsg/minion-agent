package collectors

import (
	"os"
	"runtime"
	"strings"
)

type SystemInfo struct {
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Kernel       string `json:"kernel"`
	Architecture string `json:"architecture"`
	Uptime       string `json:"uptime"`
}

func GetSystem() (*SystemInfo, error) {
	hostname, _ := os.Hostname()
	info := &SystemInfo{
		Hostname:     hostname,
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
	}

	if data, err := os.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
		info.Kernel = strings.TrimSpace(string(data))
	}

	if data, err := os.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) > 0 {
			info.Uptime = fields[0]
		}
	}

	return info, nil
}
