package internalcmd

import (
	"net"
	"push-msg/modules/logger"
	"push-msg/server"
	"time"
)

// ForwardCmd do login
type ForwardCmd struct {
}

const (
	// DefaultWriteTimeout is default timeout for write
	DefaultWriteTimeout = time.Second * 5
)

// Call implements CmdHandler
func (cmd *ForwardCmd) Call(param *server.CmdParam) (interface{}, error) {
	server := param.Server
	selfConn := param.Conn
	server.Walk(func(conn net.Conn, writeChan chan []byte) bool {
		if selfConn != conn {
			select {
			case writeChan <- []byte("hello world"):
			case <-time.After(DefaultWriteTimeout):
				logger.Error("timeout happend when writing writeChan")
			}
		}

		return true
	})

	return nil, nil
}
