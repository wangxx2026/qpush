package internalcmd

import (
	"encoding/json"
	"net"
	"push-msg/modules/logger"
	"push-msg/server"
	"push-msg/server/impl"
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
func (cmd *PushCmd) Call(param *server.CmdParam) (interface{}, error) {
	server := param.Server
	selfConn := param.Conn
	message, err := json.Marshal(param.Param)
	if err != nil {
		return false, err
	}
	packet := impl.MakePacket(param.RequestID, message)
	server.Walk(func(conn net.Conn, writeChan chan []byte) bool {
		if selfConn != conn {
			select {
			case writeChan <- packet:
			case <-time.After(DefaultWriteTimeout):
				logger.Error("timeout happend when writing writeChan")
			}
		}

		return true
	})

	return true, nil
}
