package server

import "net"

// Server is interface for server
type Server interface {
	ListenAndServe(address string, internalAddress string) error
	Walk(f func(net.Conn, chan []byte) bool)
	GetCtx(net.Conn) *ConnectionCtx
}

// Config is config for Server
type Config struct {
	ReadBufferSize int
	Handler        Handler
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
)

// CmdParam wraps param for cmd
type CmdParam struct {
	Param     []byte
	Conn      net.Conn
	Ctx       *ConnectionCtx
	Server    Server
	RequestID uint64
}

// Handler is handle for Server
type Handler interface {
	Call(cmd Cmd, internal bool, param *CmdParam) (Cmd, interface{}, error)

	RegisterCmd(cmd Cmd, internal bool, cmdHandler CmdHandler)
}

// CmdHandler is handler for cmd
type CmdHandler interface {
	Call(param *CmdParam) (Cmd, interface{}, error)
}

// ConnectionCtx is the context for connection
type ConnectionCtx struct {
	Internal bool
	GUID     []byte
}

const (
	// DefaultReadBufferSize is default read buffer size
	DefaultReadBufferSize = 10 * 1024 * 1024 // 10M
)
