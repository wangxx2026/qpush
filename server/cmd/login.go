package cmd

import (
	"encoding/json"
	"errors"
	"qpush/client"
	"qpush/modules/http"
	"qpush/modules/logger"
	"qpush/server"
)

var (
	errLoginInValidParam = errors.New("invalid param for Login cmd")
	errOfflineMsg        = errors.New("invalid offline message")
)

const (
	// DefaultReadTimeoutAfterLogin is the read timeout after login
	DefaultReadTimeoutAfterLogin = 10 * 60
)

// LoginCmd do login
type LoginCmd struct {
}

// Msg is model for message
type Msg struct {
	MsgID        int    `json:"msg_id"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Transmission string `json:"transmission"`
	Unfold       string `json:"unfold"`
}

// OfflineMsgData is data part
type OfflineMsgData struct {
	Alias   string `json:"alias"`
	MsgList []Msg  `json:"msg_list"`
}

// OfflineMsg is model for offline message
type OfflineMsg struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data OfflineMsgData `json:"data"`
}

// Call implements CmdHandler
func (cmd *LoginCmd) Call(param *server.CmdParam) (server.Cmd, interface{}, error) {
	logger.Debug("LoginCmd called")

	loginCmd := client.LoginCmd{}
	err := json.Unmarshal(param.Param, &loginCmd)
	if err != nil {
		logger.Error(errLoginInValidParam, string(param.Param))
		return server.ErrorCmd, nil, errLoginInValidParam
	}

	param.Ctx.GUID = loginCmd.GUID
	param.Ctx.AppID = loginCmd.AppID
	param.Server.BindAppGUIDToConn(loginCmd.AppID, param.Ctx.GUID, param.Conn)

	// enlarge read timeout after client login
	param.Reader.SetReadTimeout(DefaultReadTimeoutAfterLogin)

	// fetch offline messages
	data := map[string]interface{}{"app_id": loginCmd.AppID, "app_key": loginCmd.AppKey, "guid": loginCmd.GUID}
	resp, err := http.DoAkSkRequest(http.PostMethod, "/v1/pushaksk/offlinemsg", data)
	if err != nil {
		return server.ErrorCmd, nil, err
	}

	var result OfflineMsg
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return server.ErrorCmd, nil, err
	}

	alias := result.Data.Alias
	if alias != "" {
		param.Ctx.Alias = alias
	}

	return server.LoginRespCmd, result.Data.MsgList, nil
}
