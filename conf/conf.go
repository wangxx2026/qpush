package conf

import (
	"errors"
	"fmt"
	"push-msg/modules/logger"
	"sync"
)

// ServerConfig is config for Server
type ServerConfig struct {
	ReadBufferSize int
	Handler        ServerHandler
}

// CmdParam wraps param for cmd
type CmdParam struct {
	param  map[string]interface{}
	server interface{}
}

// ServerHandler is handle for Server
type ServerHandler interface {
	Call(cmd string, internal bool, param *CmdParam) (interface{}, error)

	RegisterCmd(cmd string, internal bool, cmdHandler CmdHandler)
}

// CmdHandler is handler for cmd
type CmdHandler interface {
	Call(param *CmdParam) (interface{}, error)
}

type serverHandler struct {
	cmdHandlers         sync.Map
	internalCmdHandlers sync.Map
}

// Call is call by server
func (h *serverHandler) Call(cmd string, internal bool, param *CmdParam) (interface{}, error) {
	var (
		cmdHandler interface{}
		ok         bool
	)
	if internal {
		cmdHandler, ok = h.internalCmdHandlers.Load(cmd)
	} else {
		cmdHandler, ok = h.cmdHandlers.Load(cmd)
	}
	if !ok {
		logger.Error(fmt.Printf("cmd not exists:%s", cmd))
		return nil, errCmdNotExists
	}
	return cmdHandler.(CmdHandler).Call(param)
}

// RegisterCmd registers cmd handler
func (h *serverHandler) RegisterCmd(cmd string, internal bool, cmdHandler CmdHandler) {
	if internal {
		h.internalCmdHandlers.Store(cmd, cmdHandler)
	} else {
		h.cmdHandlers.Store(cmd, cmdHandler)
	}
}

const (
	// DefaultReadBufferSize is default read buffer size
	DefaultReadBufferSize = 10 * 1024 * 1024 // 10M
)

var (
	// DefaultServerHandler is default handler
	DefaultServerHandler *serverHandler
	errCmdNotExists      = errors.New("cmd not exists")
)

func init() {
	DefaultServerHandler = &serverHandler{}
}
