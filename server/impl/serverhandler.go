package impl

import (
	"errors"
	"fmt"
	"qpush/modules/logger"
	"qpush/server"
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
func (h *ServerHandler) Call(cmd server.Cmd, internal bool, param *server.CmdParam) (server.Cmd, interface{}, error) {
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
		logger.Error(fmt.Sprintf("cmd not exists:%v", cmd))
		return server.NoCmd, nil, errCmdNotExists
	}
	return cmdHandler.(server.CmdHandler).Call(param)
}

// RegisterCmd registers cmd handler
func (h *ServerHandler) RegisterCmd(cmd server.Cmd, internal bool, cmdHandler server.CmdHandler) {
	if internal {
		h.internalCmdHandlers.Store(cmd, cmdHandler)
	} else {
		h.cmdHandlers.Store(cmd, cmdHandler)
	}
}

// Walk iterates over each handler
func (h *ServerHandler) Walk(f func(cmd server.Cmd, internal bool, cmdHandler server.CmdHandler) bool) {

	var goon bool
	h.cmdHandlers.Range(func(key, value interface{}) bool {
		goon = f(key.(server.Cmd), false, value.(server.CmdHandler))
		return goon
	})
	if !goon {
		return
	}

	h.cmdHandlers.Range(func(key, value interface{}) bool {
		goon = f(key.(server.Cmd), true, value.(server.CmdHandler))
		return goon
	})
}
