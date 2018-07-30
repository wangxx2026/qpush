package cmd

import (
	"encoding/json"
	"errors"
	"qpush/client"
	"qpush/modules/http"
	"qpush/modules/logger"
	"qpush/server"

	"github.com/zhiqiangxu/qrpc"
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

// OfflineMsgData is data part
type OfflineMsgData struct {
	Alias   string       `json:"alias"`
	MsgList []client.Msg `json:"msg_list"`
}

// OfflineMsg is model for offline message
type OfflineMsg struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data OfflineMsgData `json:"data"`
}

// ServeQRPC implements qrpc.Handler
func (cmd *LoginCmd) ServeQRPC(writer qrpc.FrameWriter, frame *qrpc.RequestFrame) {
	logger.Debug("LoginCmd called")

	jsonwriter := server.JSONFrameWriter{FrameWriter: writer}

	loginCmd := client.LoginCmd{}
	err := json.Unmarshal(frame.Payload, &loginCmd)
	if err != nil {
		logger.Error(errLoginInValidParam, string(frame.Payload))
		frame.Close()
		return
	}

	deviceInfo := &server.DeviceInfo{GUID: loginCmd.GUID, AppID: loginCmd.AppID}

	logger.Debug("test2")

	ci := frame.ConnectionInfo()
	serveconn := ci.SC
	logger.Debug(server.GetAppGUID(loginCmd.AppID, loginCmd.GUID))
	serveconn.SetID(server.GetAppGUID(loginCmd.AppID, loginCmd.GUID))
	logger.Debug("test2.5")
	// enlarge read timeout after client login
	serveconn.Reader().SetReadTimeout(DefaultReadTimeoutAfterLogin)

	logger.Debug("test3")
	// fetch offline messages
	data := map[string]interface{}{"app_id": loginCmd.AppID, "app_key": loginCmd.AppKey, "guid": loginCmd.GUID}
	resp, err := http.DoAkSkRequest(http.PostMethod, "/v1/pushaksk/offlinemsg", data)
	if err != nil {
		logger.Error("DoAkSkRequest", err)
		frame.Close()
		return
	}

	logger.Debug("test4")

	var result OfflineMsg
	err = json.Unmarshal(resp, &result)
	if err != nil {
		logger.Error("OfflineMsg Unmarshal", err)
		frame.Close()
		return
	}

	deviceInfo.Alias = result.Data.Alias

	jsonwriter.StartWrite(frame.RequestID, server.LoginRespCmd, 0)
	jsonwriter.WriteJSON(result.Data.MsgList)
	err = jsonwriter.EndWrite()
	logger.Debug("test5")
	if err != nil {
		logger.Error("EndWrite", err)
		return
	}
	logger.Debug("test6")

	ci.Anything = deviceInfo

}
