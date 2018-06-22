package client

// Agent is only different from client when dial
type Agent interface {
	Dial(address string) MsgConnection
}

// PushCmd is struct for push
type PushCmd struct {
	MsgID   int    `json:"msg_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// KillCmd is struct for kill
type KillCmd struct {
	GUID string `json:"guid"`
}
