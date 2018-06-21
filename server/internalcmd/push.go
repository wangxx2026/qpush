package internalcmd

import (
	"net"
	"qpush/modules/logger"
	"qpush/server"
	"qpush/server/impl"
	"time"
)

// PushCmd do login
type PushCmd struct {
}

const (
	// DefaultWriteTimeout is default timeout for write
	DefaultWriteTimeout = time.Second * 5
)

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
			case <-time.After(DefaultWriteTimeout):
				logger.Error("timeout happend when writing writeChan")
			}
		}

		return true
	})

	return server.PushRespCmd, true, nil
}
