package client

import "qpush/server"

// OnResponse is interface for client callback
type OnResponse interface {
	Call(uint64, server.Cmd, []byte) error
}

// Client is responsible for make connections to server
type Client interface {
	Dial(address string, guid string) MsgConnection
}

// MsgConnection is a connection to server
type MsgConnection interface {
	SendCmd(cmd string, cmdParam interface{}) (uint64, error)
	Subscribe(cb OnResponse) error
}
