package elton

import ()

type Server struct {
}

func NewServer(conf Config) (*Server, error) {
	return &Server{}, nil
}
