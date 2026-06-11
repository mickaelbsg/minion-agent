package collectors

type LoginEvent struct {
	User      string `json:"user"`
	IP        string `json:"ip"`
	Success   bool   `json:"success"`
	Timestamp string `json:"timestamp"`
}

func GetLogins() ([]LoginEvent, error) {
	return []LoginEvent{}, nil
}
