package collectors

import "os/user"

type User struct {
	Username string `json:"username"`
	UID      string `json:"uid"`
	GID      string `json:"gid"`
	Home     string `json:"home"`
}

func GetUsers() ([]User, error) {
	current, err := user.Current()
	if err != nil {
		return nil, err
	}

	return []User{
		{
			Username: current.Username,
			UID:      current.Uid,
			GID:      current.Gid,
			Home:     current.HomeDir,
		},
	}, nil
}
