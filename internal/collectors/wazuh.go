package collectors

type WazuhStatus struct {
	Connected bool   `json:"connected"`
	Version   string `json:"version"`
	LastSeen  string `json:"last_seen"`
}

func GetWazuhStatus() (*WazuhStatus, error) {
	return &WazuhStatus{}, nil
}
