package impl

import (
	"errors"
	"fmt"
	"push-msg/modules/logger"
	"push-msg/server"
	"sync"
)

var (
	errCmdNotExists = errors.New("cmd not exists")
)

// ServerHandler is the implement
type ServerHandler struct {
	cmdHandlers         sync.Map
	internalCmdHandlers sync.Map
}

// Call is call by server
func (h *ServerHandler) Call(cmd string, internal bool, param *server.CmdParam) (interface{}, error) {
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
		logger.Error(fmt.Sprintf("cmd not exists:%s", cmd))
		return nil, errCmdNotExists
	}
	return cmdHandler.(server.CmdHandler).Call(param)
}

// RegisterCmd registers cmd handler
func (h *ServerHandler) RegisterCmd(cmd string, internal bool, cmdHandler server.CmdHandler) {
	if internal {
		h.internalCmdHandlers.Store(cmd, cmdHandler)
	} else {
		h.cmdHandlers.Store(cmd, cmdHandler)
	}
}
