package internalcmd

import "qpush/server"

// StatusCmd do status
type StatusCmd struct {
}

// Call implements CmdHandler
func (cmd *StatusCmd) Call(param *server.CmdParam) (server.Cmd, interface{}, error) {

	status := param.Server.GetStatus()
	return server.StatusRespCmd, status, nil

}
