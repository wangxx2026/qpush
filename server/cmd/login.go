package cmd

import (
	"push-msg/conf"
)

// LoginCmd do login
type LoginCmd struct {
}

// Call implements CmdHandler
func (cmd *LoginCmd) Call(param *conf.CmdParam) (interface{}, error) {
	return "hello", nil
}
