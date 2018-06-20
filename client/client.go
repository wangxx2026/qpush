package client

import (
	"push-msg/client/impl"
)

// New creates a Client instance
func New() Client {
	return impl.NewClient()
}

// OnResponse is interface for client callback
type OnResponse interface {
	Call(msg string) error
}

// Client is responsible for make connections to server
type Client interface {
	Dial(address string, guid string) MsgConnection
}

// MsgConnection is a connection to server
type MsgConnection interface {
	Send(msg string) error
	Subscribe(cb OnResponse) error
}
