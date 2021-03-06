package cmd

import (
	"encoding/json"
	"errors"
	"qpush/client"
	"qpush/pkg/http"
	"qpush/pkg/logger"
	"qpush/server"
	"sync"
	"time"
	"unsafe"

	"github.com/zhiqiangxu/qrpc"
)

// NewAckCmd creates an AckCmd instance
func NewAckCmd() *AckCmd {
	cmd := &AckCmd{queuedAck: make(map[string]map[string]map[int]struct{}), batchSignal: make(chan struct{}, 1)}
	go cmd.syncAck()
	return cmd
}

// AckCmd do ack
type AckCmd struct {
	lock        sync.Mutex
	batchSignal chan struct{}
	queuedAck   map[string]map[string]map[int]struct{} //appguid msg_id type
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

// ServeQRPC implements qrpc.Handler
func (cmd *AckCmd) ServeQRPC(writer qrpc.FrameWriter, frame *qrpc.RequestFrame) {

	ci := frame.ConnectionInfo()
	anything := ci.GetAnything()
	deviceInfo, ok := anything.(*server.DeviceInfo)
	if !ok {
		logger.Error("failed to cast DeviceInfo\n")
		frame.Close()
		return
	}

	if deviceInfo.GUID == "" {
		logger.Error(errAckNoGUID)
		frame.Close()
		return
	}
	appGUID := server.GetAppGUID(deviceInfo.AppID, deviceInfo.GUID)
	logger.Info("AckCmdLOG", string(frame.Payload), appGUID)

	serveconn := ci.SC
	mem := unsafe.Pointer(serveconn)
	logger.Debug(mem, string(frame.Payload))

	ackCmd := client.AckCmd{}
	err := json.Unmarshal(frame.Payload, &ackCmd)
	if err != nil || len(ackCmd.MsgIDS) == 0 {
		logger.Error(errAckInValidParam)
		frame.Close()
		return
	}

	cmd.lock.Lock()

	var emptyID bool
	ack, ok := cmd.queuedAck[appGUID]
	if !ok {
		ack = make(map[string]map[int]struct{})
		cmd.queuedAck[appGUID] = ack
	}
	for _, id := range ackCmd.MsgIDS {
		if id == "" {
			emptyID = true
		}
		_, ok := ack[id]
		if !ok {
			ack[id] = make(map[int]struct{})
		}
		ack[id][ackCmd.Type] = struct{}{}
	}

	if len(cmd.queuedAck) >= BatchAckNumber {
		select {
		case cmd.batchSignal <- struct{}{}:
		default:
		}
	}

	cmd.lock.Unlock()
	if emptyID {
		logger.Error("emptyId", appGUID, ackCmd)
	}
	logger.Debug("AckCmd called")
	jsonBytes, err := json.Marshal(true)
	if err != nil {
		logger.Error("Marshal true failed", err)
		frame.Close()
		return
	}
	writer.StartWrite(frame.RequestID, server.AckRespCmd, 0)
	writer.WriteBytes(jsonBytes)
	err = writer.EndWrite()
	if err != nil {
		logger.Error("EndWrite", err)
		return
	}
	// logger.Info("qrpc request param", string(frame.Payload))
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
		cmd.queuedAck = make(map[string]map[string]map[int]struct{})
		cmd.lock.Unlock()

		ackData := make(map[string]map[string][]int)
		for appGUID, idMap := range queuedAck {
			ids := make(map[string][]int)
			for id, typeMap := range idMap {
				types := make([]int, 0, len(typeMap))
				for tp := range typeMap {
					types = append(types, tp)
				}
				ids[id] = types
			}
			ackData[appGUID] = ids

			if len(ackData) > BatchAckNumber {
				cmd.syncBatch(ackData)
				ackData = make(map[string]map[string][]int)
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
type idType struct {
	MsgID string `json:"msg_id"`
	Types []int  `json:"types"`
}
type ackRecord struct {
	IDTypes []idType `json:"id_types"`
	GUID    string   `json:"guid"`
}
type ackResponse struct {
	Code int    `json:"code"`
	MSG  string `json:"msg"`
}

func (cmd *AckCmd) syncBatch(ackData map[string]map[string][]int) {
	request := ackRequest{NotifyData: make([]ackRecord, 0, len(ackData))}
	for appGUID, id2Types := range ackData {
		idTypes := make([]idType, 0, len(id2Types))
		for id, types := range id2Types {
			idTypes = append(idTypes, idType{MsgID: id, Types: types})
		}
		record := ackRecord{IDTypes: idTypes, GUID: appGUID}
		request.NotifyData = append(request.NotifyData, record)
	}
	// bytes, _ := json.Marshal(request)
	// logger.Info("http request param", string(bytes))

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
