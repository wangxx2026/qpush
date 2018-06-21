package cmd

import (
	"encoding/json"
	"errors"
	"qpush/client"
	"qpush/modules/logger"
	"qpush/server"
)

var (
	errLoginInValidParam = errors.New("invalid param for Login cmd")
)

// LoginCmd do login
type LoginCmd struct {
}

// Call implements CmdHandler
func (cmd *LoginCmd) Call(param *server.CmdParam) (server.Cmd, interface{}, error) {
	logger.Info("LoginCmd called")

	loginCmd := client.LoginCmd{}
	err := json.Unmarshal(param.Param, &loginCmd)
	if err != nil {
		logger.Error(errLoginInValidParam)
		return server.ErrorCmd, nil, errLoginInValidParam
	}

	param.Ctx.GUID = []byte(loginCmd.GUID)

	// TODO fetch offline messages
	// client := http.Client{Timeout: 3*time.Second}
	// resp, err := client.Get("http://example.com")

	return server.LoginRespCmd, true, nil
}
