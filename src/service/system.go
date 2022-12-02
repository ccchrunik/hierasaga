package service

type System struct {
	Services map[string]Service
}

func NewSystem() *System {
	return &System{}
}
