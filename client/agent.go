package client

// Agent is only different from client when dial
type Agent interface {
	Dial(address string) MsgConnection
}

const (
	// PushCmdName is name of push cmd
	PushCmdName = "push"
)

// PushCmd is struct for push
type PushCmd struct {
	MsgID   int    `json:"msg_id"`
	Message string `json:"message"`
}

// AckCmd is for ack of message
type AckCmd struct {
	MsgID int `json:"msg_id"`
}
