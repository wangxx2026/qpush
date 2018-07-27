package client

// Msg is model for message
type Msg struct {
	MsgID        string `json:"msg_id"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Transmission string `json:"transmission"`
	Unfold       string `json:"unfold"`
	PassThrough  int    `json:"pass_through"`
}

// PushCmd is struct for push
type PushCmd struct {
	Msg   `json:"msg"`
	AppID int      `json:"app_id"`
	GUID  []string `json:"guid"`
}

// KillCmd is struct for kill
type KillCmd struct {
	GUID  string `json:"guid"`
	AppID int    `json:"app_id"`
}

// ExecCmd is struct for exec
type ExecCmd struct {
	Cmd string
}

// ExecRespCmd is struct for execresp
type ExecRespCmd struct {
	Err string
}
