package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"qpush/client"
	"qpush/modules/http"
	"qpush/modules/logger"
	"qpush/server"
	"sync"
	"time"
)

// NewAckCmd creates an AckCmd instance
func NewAckCmd() *AckCmd {
	cmd := &AckCmd{queuedAck: make(map[string]map[string]bool), batchSignal: make(chan bool, 1)}
	go cmd.syncAck()
	return cmd
}

// AckCmd do ack
type AckCmd struct {
	lock        sync.Mutex
	batchSignal chan bool
	queuedAck   map[string]map[string]bool
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
		cmd.queuedAck[appGUID] = make(map[string]bool)
	}
	for _, id := range ackCmd.MsgIDS {
		cmd.queuedAck[appGUID][id] = true
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
		cmd.queuedAck = make(map[string]map[string]bool)
		cmd.lock.Unlock()

		ackData := make(map[string][]string)
		for appGUID, idMap := range queuedAck {
			ids := make([]string, 0, len(idMap))
			for id := range idMap {
				ids = append(ids, id)
			}

			ackData[appGUID] = ids

			if len(ackData) > BatchAckNumber || len(ids) > BatchAckNumber {
				cmd.syncBatch(ackData)
				ackData = make(map[string][]string)
			}
		}

		if len(ackData) > 0 {
			cmd.syncBatch(ackData)
		}

		ackData = nil // maybe good for gc

	}

}

type ackRequest struct {
	NotifyData []ackRecord `json:"notify_data"`
}
type ackRecord struct {
	MsgIDS []string `json:"msg_ids"`
	GUID   string   `json:"guid"`
}
type ackResponse struct {
	Code int    `json:"code"`
	MSG  string `json:"msg"`
}

func (cmd *AckCmd) syncBatch(ackData map[string][]string) {
	request := ackRequest{NotifyData: make([]ackRecord, 0, len(ackData))}
	for appGUID, ids := range ackData {
		record := ackRecord{MsgIDS: ids, GUID: appGUID}
		request.NotifyData = append(request.NotifyData, record)
	}

	for {
		// send ack message
		resp, err := http.DoAkSkRequest(http.PostMethod, "/v1/pushaksk/notifymsg", &request)
		if err != nil {
			logger.Error("error in DoAkSkRequest", err)
			time.Sleep(time.Second)
			continue
		}

		var result ackResponse
		err = json.Unmarshal(resp, &result)
		if err != nil {
			logger.Error("error in DoAkSkRequest response", err)
			time.Sleep(time.Second)
			continue
		}
		if result.Code == 0 {
			return
		}

		logger.Error("error in result.Code", result.Code)
		time.Sleep(time.Second)

	}

}

// Status returns status of this cmd
func (cmd *AckCmd) Status() interface{} {
	cmd.lock.Lock()

	var count int
	for _, idMap := range cmd.queuedAck {
		count += len(idMap)
	}

	cmd.lock.Unlock()

	return map[string]int{"queuedAck": count}
}
