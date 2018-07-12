package server

import (
	"encoding/binary"
	"errors"
	"net"
	"time"
)

var (
	// ErrMarshalFail for marshal fail
	ErrMarshalFail = errors.New("failed to marshal")
	// ErrUnMarshalFail for unmarshal fail
	ErrUnMarshalFail = errors.New("failed to unmarshal")
	// ErrInvalidParam when param not valid
	ErrInvalidParam = errors.New("invalid param")
	// ErrConnectionClosed for connection closed
	ErrConnectionClosed = errors.New("connection closed")
)

// Server is interface for server
type Server interface {
	ListenAndServe(address string, internalAddress string) error
	Walk(f func(net.Conn, *ConnectionCtx) bool)
	GetStatus() *Status
	BindAppGUIDToConn(int, string, net.Conn)
	SendTo(int, []string, []byte) int
	KillAppGUID(appID int, guid string) error
	CloseConnection(conn net.Conn) (<-chan bool, error)
}

// Config is config for Server
type Config struct {
	ReadBufferSize int
	Handler        Handler
	HBConfig       HeartBeatConfig
}

// HeartBeatConfig if config for heartbeat
type HeartBeatConfig struct {
	Callback func() error
	Interval time.Duration
}

// Cmd is uint32
type Cmd uint32

const (
	// LoginCmd is for outside
	LoginCmd Cmd = iota
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
)

// CmdParam wraps param for cmd
type CmdParam struct {
	Param     []byte
	Conn      net.Conn
	Reader    StreamReader
	Ctx       *ConnectionCtx
	Server    Server
	RequestID uint64
}

// CmdCall for call cmd
type CmdCall struct {
	Cmd   Cmd
	Param *CmdParam
}

// StreamReader is interface to stream reader
type StreamReader interface {
	SetReadTimeout(timeout int)
}

// Handler is handle for Server
type Handler interface {
	Call(cmd Cmd, internal bool, param *CmdParam) (Cmd, interface{}, error)

	RegisterCmd(cmd Cmd, internal bool, cmdHandler CmdHandler)

	Walk(func(cmd Cmd, internal bool, cmdHandler CmdHandler) bool)
}

// CmdHandler is handler for cmd
type CmdHandler interface {
	Call(param *CmdParam) (Cmd, interface{}, error)
}

// StatusAble is interface for status
type StatusAble interface {
	Status() interface{}
}

// ConnectionCtx is the context for connection
type ConnectionCtx struct {
	Internal  bool
	GUID      string
	AppID     int
	WriteChan chan []byte
	CloseChan chan bool // only close this channel
}

// Status contains server status info
type Status struct {
	GUIDCount       int
	GUIDConnMapSize int
	ConnCtxMapSize  int
	HandleStatus    map[Cmd]interface{}
	Uptime          time.Time
}

const (
	// DefaultReadBufferSize is default read buffer size
	DefaultReadBufferSize = 10 * 1024 * 1024 // 10M
)

// MakePacket generate packet
func MakePacket(requestID uint64, cmd Cmd, payload []byte) []byte {
	length := 12 + uint32(len(payload))
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf, length)
	binary.BigEndian.PutUint64(buf[4:], requestID)
	binary.BigEndian.PutUint32(buf[12:], uint32(cmd))
	copy(buf[16:], payload)
	return buf
}
