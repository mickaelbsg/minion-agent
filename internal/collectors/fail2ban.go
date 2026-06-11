package collectors

type Fail2BanEvent struct {
	IP        string `json:"ip"`
	Jail      string `json:"jail"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
}

func GetFail2BanEvents() ([]Fail2BanEvent, error) {
	return []Fail2BanEvent{}, nil
}
