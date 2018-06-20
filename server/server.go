package server

// Server is interface for server
type Server interface {
	ListenAndServe(address string, internalAddress string) error
}

// Config is config for Server
type Config struct {
	ReadBufferSize int
	Handler        Handler
}

// CmdParam wraps param for cmd
type CmdParam struct {
	param  map[string]interface{}
	server interface{}
}

// ServerHandler is handle for Server
type Handler interface {
	Call(cmd string, internal bool, param *CmdParam) (interface{}, error)

	RegisterCmd(cmd string, internal bool, cmdHandler CmdHandler)
}

// CmdHandler is handler for cmd
type CmdHandler interface {
	Call(param *CmdParam) (interface{}, error)
}

const (
	// DefaultReadBufferSize is default read buffer size
	DefaultReadBufferSize = 10 * 1024 * 1024 // 10M
)
