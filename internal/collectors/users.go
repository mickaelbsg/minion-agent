package collectors

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type User struct {
	Username string `json:"username"`
	UID      string `json:"uid"`
	GID      string `json:"gid"`
	Home     string `json:"home"`
}

func GetUsers() ([]User, error) {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	users := []User{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) < 7 {
			continue
		}

		username := fields[0]
		uid := fields[2]
		gid := fields[3]
		home := fields[5]
		shell := fields[6]

		if !isHumanUser(uid, shell) {
			continue
		}

		users = append(users, User{
			Username: username,
			UID:      uid,
			GID:      gid,
			Home:     home,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func isHumanUser(uidValue string, shell string) bool {
	uid, err := strconv.Atoi(uidValue)
	if err != nil {
		return false
	}

	if strings.HasSuffix(shell, "/nologin") || strings.HasSuffix(shell, "/false") {
		return false
	}

	return uid == 0 || uid >= 1000
}
