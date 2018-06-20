package server

import (
	"push-msg/conf"
	"push-msg/server/impl"
)

// New creates a Server instance
func New(conf *conf.ServerConfig) Server {
	return impl.NewServer(conf)
}

// Server is interface for server
type Server interface {
	ListenAndServe(address string, internalAddress string) error
}
