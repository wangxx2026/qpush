package internalcmd

import (
	"encoding/json"
	"errors"
	"qpush/client"
	"qpush/modules/logger"
	"qpush/server"
)

var (
	errUnmarshalFail = errors.New("Unmarshal failed")
)

// KillCmd do kill
type KillCmd struct {
}

// Call implements CmdHandler
func (cmd *KillCmd) Call(param *server.CmdParam) (server.Cmd, interface{}, error) {

	cmdInfo := &client.KillCmd{}
	err := json.Unmarshal(param.Param, cmdInfo)
	if err != nil {
		logger.Error(errUnmarshalFail, err)
		return server.ErrorCmd, nil, errUnmarshalFail
	}

	err = param.Server.KillAppGUID(cmdInfo.AppID, cmdInfo.GUID)

	return server.KillRespCmd, true, err

}
