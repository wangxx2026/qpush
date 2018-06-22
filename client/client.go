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
	SendCmd(cmd server.Cmd, cmdParam interface{}) (uint64, error)
	SendCmdBlocking(cmd server.Cmd, cmdParam interface{}) ([]byte, error)
	Subscribe(cb OnResponse) error
}

// LoginCmd is for login
type LoginCmd struct {
	GUID string `json:"guid"`
}

// AckCmd is for ack of message
type AckCmd struct {
	MsgID int `json:"msg_id"`
}
