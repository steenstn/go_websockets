package main

type ServerFullError struct {
}

func (m *ServerFullError) Error() string {
	return "No free client slots"
}
