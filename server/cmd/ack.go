package cmd

import (
	"encoding/json"
	"errors"
	"net/url"
	"qpush/client"
	"qpush/modules/logger"
	"qpush/server"
	"strconv"
	"sync"
	"time"
)

// NewAckCmd creates an AckCmd instance
func NewAckCmd() *AckCmd {
	cmd := &AckCmd{queuedAck: make(map[string]int), batchSignal: make(chan bool, 10)}
	go cmd.syncAck()
	return cmd
}

// AckCmd do ack
type AckCmd struct {
	lock        sync.Mutex
	batchSignal chan bool
	queuedAck   map[string]int
}

const (
	// BatchAckNumber is threshold for batch ack
	BatchAckNumber = 100
	// BatchAckTimeout guarantees at least once per 60s
	BatchAckTimeout = 60 * time.Second
)

var (
	errAckNoGUID       = errors.New("ack with no guid")
	errAckInValidParam = errors.New("Ack with invalid param")
)

// Call implements CmdHandler
func (cmd *AckCmd) Call(param *server.CmdParam) (server.Cmd, interface{}, error) {

	if param.Ctx.GUID == "" {
		logger.Error(errAckNoGUID)
		return server.ErrorCmd, nil, errAckNoGUID
	}
	guid := param.Ctx.GUID

	ackCmd := client.AckCmd{}
	err := json.Unmarshal(param.Param, &ackCmd)
	if err != nil {
		logger.Error(errAckInValidParam)
		return server.ErrorCmd, nil, errAckInValidParam
	}

	cmd.lock.Lock()
	defer cmd.lock.Unlock()

	offset, ok := cmd.queuedAck[guid]
	if !ok || ackCmd.MsgID > offset {
		cmd.queuedAck[guid] = ackCmd.MsgID
	}

	if len(cmd.queuedAck) >= BatchAckNumber {
		cmd.batchSignal <- true
	}

	logger.Info("AckCmd called")
	return server.AckRespCmd, true, nil
}

func (cmd *AckCmd) syncAck() {
	for {
		select {
		case <-cmd.batchSignal:
		case <-time.After(BatchAckTimeout):
		}
		cmd.lock.Lock()

		ackData := make(url.Values)
		for guid, offset := range cmd.queuedAck {
			ackData.Set(guid, strconv.Itoa(offset))
		}

		cmd.queuedAck = make(map[string]int)
		cmd.lock.Unlock()

		// TODO send ackData

	}
}
