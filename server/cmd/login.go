package cmd

import (
	"push-msg/modules/logger"
	"push-msg/server"
)

// LoginCmd do login
type LoginCmd struct {
}

// Call implements CmdHandler
func (cmd *LoginCmd) Call(param *server.CmdParam) (server.Cmd, interface{}, error) {
	logger.Info("LoginCmd called")
	return server.LoginRespCmd, "hello", nil
}
