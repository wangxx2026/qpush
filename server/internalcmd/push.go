package internalcmd

import (
	"net"
	"qpush/modules/logger"
	"qpush/server"
	"qpush/server/impl"
)

// PushCmd do login
type PushCmd struct {
}

// Call implements CmdHandler
func (cmd *PushCmd) Call(param *server.CmdParam) (server.Cmd, interface{}, error) {
	s := param.Server
	selfConn := param.Conn
	message := param.Param

	packet := impl.MakePacket(param.RequestID, server.ForwardCmd, message)
	s.Walk(func(conn net.Conn, writeChan chan []byte) bool {
		if selfConn != conn {
			ctx := s.GetCtx(conn)
			if ctx.Internal {
				return true
			}
			select {
			case writeChan <- packet:
			default:
				logger.Error("writeChan blocked for", ctx.GUID)
			}
		}

		return true
	})

	return server.PushRespCmd, true, nil
}
