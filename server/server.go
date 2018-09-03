package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zhiqiangxu/qrpc"
)

var (
	// ErrMarshalFail for marshal fail
	ErrMarshalFail = errors.New("failed to marshal")
	// ErrUnMarshalFail for unmarshal fail
	ErrUnMarshalFail = errors.New("failed to unmarshal")
	// ErrInvalidParam when param not valid
	ErrInvalidParam = errors.New("invalid param")
	// ErrCanceled when canceled
	ErrCanceled = errors.New("canceled")
	// ErrConnectionClosed for connection closed
	ErrConnectionClosed = errors.New("connection closed")
)

const (
	// LoginCmd is for outside
	LoginCmd qrpc.Cmd = iota
	// LoginRespCmd is resp for login
	LoginRespCmd
	// PushCmd is for internal
	PushCmd
	// PushRespCmd is resp for push
	PushRespCmd
	// ForwardCmd is cmd when do forwarding
	ForwardCmd
	// NoCmd is like 404 for http
	NoCmd
	// AckCmd is for ack msg
	AckCmd
	// AckRespCmd is resp for ack
	AckRespCmd
	// ErrorCmd is when resp error
	ErrorCmd
	// HeartBeatCmd is for keep alive
	HeartBeatCmd
	// HeartBeatRespCmd is resp for heartbeat
	HeartBeatRespCmd
	// StatusCmd is for query server status
	StatusCmd
	// StatusRespCmd is for query server status
	StatusRespCmd
	// KillCmd is for kill specific guid
	KillCmd
	// KillRespCmd is resp for KillCmd
	KillRespCmd
	// KillAllCmd is for kill all cons
	KillAllCmd
	// KillAllRespCmd is resp for KillAllCmd
	KillAllRespCmd
	// ListGUIDCmd is for list guid
	ListGUIDCmd
	// ListGUIDRespCmd is resp for ListGUIDCmd
	ListGUIDRespCmd
	// ExecCmd for exec
	ExecCmd
	// ExecRespCmd is resp for ExecCmd
	ExecRespCmd
	// RelayCmd for relay
	RelayCmd
	// RelayRespCmd is resp for RelayCmd
	RelayRespCmd
	// CheckGUIDCmd for checkguid
	CheckGUIDCmd
	// CheckGUIDRespCmd is resp for CheckGUIDCmd
	CheckGUIDRespCmd
)

// DeviceInfo defines info on connection
type DeviceInfo struct {
	Uptime time.Time
	GUID   string
	AppID  int
	CityID int
	TagIDS []int
	OS     string
}

// DeviceInfoSlice for slice of DeviceInfo
type DeviceInfoSlice []*DeviceInfo

func (s DeviceInfoSlice) Len() int           { return len(s) }
func (s DeviceInfoSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s DeviceInfoSlice) Less(i, j int) bool { return s[i].Uptime.Before(s[j].Uptime) }

// GetAppGUID creates unique id by appID and guid
func GetAppGUID(appID int, guid string) string {
	return fmt.Sprintf("%d:%s", appID, guid)
}

// JSONFrameWriter for write json
type JSONFrameWriter struct {
	qrpc.FrameWriter
}

// WriteJSON write json with FrameWriter
func (writer JSONFrameWriter) WriteJSON(value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	writer.WriteBytes(bytes)
	return nil
}
