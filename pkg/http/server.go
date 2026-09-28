package http

import (
	"net"
	"net/http"
)

type Server struct {
	http.Server
	Network string
}

func (s *Server) ListenAndServe() error {
	if s.Network == "tcp" || s.Network == "" {
		return s.Server.ListenAndServe()
	}
	l, err := net.Listen(s.Network, s.Addr)
	if err != nil {
		return err
	}
	return s.Serve(l)
}
