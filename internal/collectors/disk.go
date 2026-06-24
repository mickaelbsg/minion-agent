package collectors

import (
	"os/exec"

	"strings"
)

type DiskInfo struct {
	Filesystem string `json:"filesystem"`
	Size       string `json:"size"`
	Used       string `json:"used"`
	Avail      string `json:"avail"`
	UsePercent string `json:"use_percent"`
	MountedOn  string `json:"mounted_on"`
}

func GetDiskUsage() ([]DiskInfo, error) {
	cmd := exec.Command("df", "-h")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var disks []DiskInfo
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		disks = append(disks, DiskInfo{
			Filesystem: fields[0],
			Size:       fields[1],
			Used:       fields[2],
			Avail:      fields[3],
			UsePercent: fields[4],
			MountedOn:  fields[5],
		})
	}
	return disks, nil
}
