package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"qpush/client"
	"qpush/modules/logger"
	"qpush/server"
	"strconv"
	"strings"
	"sync"
	"time"
)

// NewAckCmd creates an AckCmd instance
func NewAckCmd() *AckCmd {
	cmd := &AckCmd{queuedAck: make(map[string]map[int]bool), batchSignal: make(chan bool, 1)}
	go cmd.syncAck()
	return cmd
}

// AckCmd do ack
type AckCmd struct {
	lock        sync.Mutex
	batchSignal chan bool
	queuedAck   map[string]map[int]bool
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
	appGUID := fmt.Sprintf("%d:%s", param.Ctx.AppID, param.Ctx.GUID)

	ackCmd := client.AckCmd{}
	err := json.Unmarshal(param.Param, &ackCmd)
	if err != nil {
		logger.Error(errAckInValidParam)
		return server.ErrorCmd, nil, errAckInValidParam
	}

	cmd.lock.Lock()
	defer cmd.lock.Unlock()

	_, ok := cmd.queuedAck[appGUID]
	if !ok {
		cmd.queuedAck[appGUID] = make(map[int]bool)
	}
	ids := strings.Split(ackCmd.MsgIDS, ",")
	for _, id := range ids {
		idInt, err := strconv.Atoi(id)
		if err != nil {
			return server.ErrorCmd, nil, errAckInValidParam
		}
		cmd.queuedAck[appGUID][idInt] = true
	}

	if len(cmd.queuedAck) >= BatchAckNumber || len(cmd.queuedAck[appGUID]) > BatchAckNumber {
		select {
		case cmd.batchSignal <- true:
		default:
		}

	}

	logger.Debug("AckCmd called")
	return server.AckRespCmd, true, nil
}

func (cmd *AckCmd) syncAck() {
	for {
		select {
		case <-cmd.batchSignal:
		case <-time.After(BatchAckTimeout):
		}
		cmd.lock.Lock()
		// fast copy and unlock
		queuedAck := cmd.queuedAck
		cmd.queuedAck = make(map[string]map[int]bool)
		cmd.lock.Unlock()

		ackData := make(map[string]string)
		for appGUID, idMap := range queuedAck {
			ids := make([]string, 0, len(idMap))
			for id := range idMap {
				ids = append(ids, string(id))
			}

			ackData[appGUID] = strings.Join(ids, ",")
		}

		// TODO send ackData

	}
}
