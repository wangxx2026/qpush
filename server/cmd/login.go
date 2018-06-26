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

const (
	// DefaultReadTimeoutAfterLogin is the read timeout after login
	DefaultReadTimeoutAfterLogin = 10 * 60
)

// LoginCmd do login
type LoginCmd struct {
}

// Call implements CmdHandler
func (cmd *LoginCmd) Call(param *server.CmdParam) (server.Cmd, interface{}, error) {
	logger.Debug("LoginCmd called")

	loginCmd := client.LoginCmd{}
	err := json.Unmarshal(param.Param, &loginCmd)
	if err != nil {
		logger.Error(errLoginInValidParam)
		return server.ErrorCmd, nil, errLoginInValidParam
	}

	param.Ctx.GUID = loginCmd.GUID
	param.Server.BindGUIDToConn(loginCmd.GUID, param.Conn)

	// enlarge read timeout after client login
	param.Reader.SetReadTimeout(DefaultReadTimeoutAfterLogin)

	// TODO fetch offline messages
	// client := http.Client{Timeout: 3*time.Second}
	// resp, err := client.Get("http://example.com")

	return server.LoginRespCmd, true, nil
}
