package client

// Msg is model for message
type Msg struct {
	Style        int    `json:"style"`
	MsgID        string `json:"msg_id"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Transmission string `json:"transmission"`
	Unfold       string `json:"unfold"`
	PassThrough  int    `json:"pass_through"`
}

// PushCmd is struct for push
type PushCmd struct {
	Msg     `json:"msg"`
	AppID   int      `json:"app_id"`
	GUID    []string `json:"guid"`
	CityIDS []int    `json:"city_ids"`
	TagIDS  []int    `json:"tag_ids"`
	OS      string   `json:"os"`
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

// CheckGUIDCmd is struct for checkguid
type CheckGUIDCmd struct {
	GUID  string `json:"guid"`
	AppID int    `json:"app_id"`
}
