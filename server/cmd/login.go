package cmd

import (
	"push-msg/server"
)

// LoginCmd do login
type LoginCmd struct {
}

// Call implements CmdHandler
func (cmd *LoginCmd) Call(param *server.CmdParam) (interface{}, error) {
	return "hello", nil
}
