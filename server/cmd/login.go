package cmd

import (
	"encoding/json"
	"errors"
	"qpush/client"
	"qpush/pkg/http"
	"qpush/pkg/logger"
	"qpush/server"
	"strconv"
	"sync"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/zhiqiangxu/qrpc"
)

var (
	// ErrLoginInValidParam when invalid param
	ErrLoginInValidParam = errors.New("invalid param for Login cmd")
)

const (
	// DefaultReadTimeoutAfterLogin is the read timeout after login
	DefaultReadTimeoutAfterLogin = 10 * 60
)

// LoginCmd do login
type LoginCmd struct {
	m             sync.Mutex
	onlineStat    map[int]int64 //appid -> count
	gaugeMetric   metrics.Gauge
	counterMetric metrics.Counter
}

// NewLoginCmd returns a LoginCmd instance
func NewLoginCmd(gaugeMetric metrics.Gauge, counterMetric metrics.Counter) *LoginCmd {
	return &LoginCmd{onlineStat: make(map[int]int64), gaugeMetric: gaugeMetric, counterMetric: counterMetric}
}

// OfflineMsgData is data part
type OfflineMsgData struct {
	CityID  int          `json:"city_id"`
	OS      string       `json:"os"`
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
		logger.Error(ErrLoginInValidParam, string(frame.Payload))
		frame.Close()
		return
	}

	deviceInfo := &server.DeviceInfo{Uptime: time.Now(), GUID: loginCmd.GUID, AppID: loginCmd.AppID}

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

	logger.Debug("resp", string(resp))

	var result OfflineMsg
	err = json.Unmarshal(resp, &result)
	if err != nil {
		logger.Error("OfflineMsg Unmarshal", err)
		frame.Close()
		return
	}
	if result.Code != 0 {
		logger.Error("login failed", string(resp))
		frame.Close()
		return
	}

	deviceInfo.CityID = result.Data.CityID
	deviceInfo.OS = result.Data.OS

	jsonwriter.StartWrite(frame.RequestID, server.LoginRespCmd, 0)
	jsonwriter.WriteJSON(result.Data.MsgList)
	err = jsonwriter.EndWrite()
	logger.Debug("test5")

	if err != nil {
		counterNGLabels := []string{"appid", strconv.Itoa(loginCmd.AppID), "kind", "offlineng"}
		cmd.counterMetric.With(counterNGLabels...).Add(1)
		logger.Error("EndWrite", err)
		return
	}

	counterOKLabels := []string{"appid", strconv.Itoa(loginCmd.AppID), "kind", "offlineok"}
	cmd.counterMetric.With(counterOKLabels...).Add(1)

	logger.Debug("test6")

	ci.SetAnything(deviceInfo)

	cmd.m.Lock()
	v := cmd.onlineStat[loginCmd.AppID]
	cmd.onlineStat[loginCmd.AppID] = v + 1
	cmd.m.Unlock()

	labels := []string{"appid", strconv.Itoa(loginCmd.AppID), "kind", "online"}
	cmd.gaugeMetric.With(labels...).Set(float64(v + 1))

	ci.NotifyWhenClose(func() {
		cmd.m.Lock()
		v := cmd.onlineStat[loginCmd.AppID]
		if v > 0 {
			cmd.onlineStat[loginCmd.AppID] = v - 1
			cmd.m.Unlock()
			cmd.gaugeMetric.With(labels...).Set(float64(v - 1))
		} else {
			cmd.m.Unlock()
			logger.Error("bug happend")
		}

	})
}
