package internalcmd

import (
	"net"
	"qpush/server"
)

// KillAllCmd do kill
type KillAllCmd struct {
}

// Call implements CmdHandler
func (cmd *KillAllCmd) Call(param *server.CmdParam) (server.Cmd, interface{}, error) {

	selfConn := param.Conn
	s := param.Server

	s.Walk(func(conn net.Conn, ctx *server.ConnectionCtx) bool {
		if selfConn != conn {
			s.CloseConnection(conn)
		}

		return true
	})
	return server.KillAllRespCmd, true, nil

}

// Status returns status of this cmd
func (cmd *KillAllCmd) Status() interface{} {
	return nil
}
