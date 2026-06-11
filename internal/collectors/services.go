package collectors

type Service struct {
	Name        string `json:"name"`
	State       string `json:"state"`
	SubState    string `json:"sub_state"`
	Description string `json:"description"`
}

func GetServices() ([]Service, error) {
	return []Service{}, nil
}
