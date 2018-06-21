package impl

import (
	"fmt"
	"net"
	"push-msg/modules/logger"
)

// Agent is data structor for client
type Agent struct {
}

// NewAgent creates a Agent instance
func NewAgent() *Agent {
	return &Agent{}
}

// Dial initiates a connection to server
func (agent *Agent) Dial(address string) *MsgConnection {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		logger.Error(fmt.Sprintf("agent failed to dial %s", address))
		return nil
	}

	mc := &MsgConnection{conn: conn}
	return mc
}
