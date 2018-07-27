package client

// LoginCmd is for login
type LoginCmd struct {
	GUID   string `json:"guid"`
	AppID  int    `json:"app_id"`
	AppKey string `json:"app_key"`
}

// AckCmd is for ack of message
type AckCmd struct {
	MsgIDS []string `json:"msg_ids"`
}
