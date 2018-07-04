package cmd

import "qpush/server"

// HeartBeatCmd do heartbeat
type HeartBeatCmd struct {
}

// Call implements CmdHandler
func (cmd *HeartBeatCmd) Call(param *server.CmdParam) (server.Cmd, interface{}, error) {
	return server.HeartBeatRespCmd, nil, nil
}

// Status returns status of this cmd
func (cmd *HeartBeatCmd) Status() interface{} {
	return nil
}
