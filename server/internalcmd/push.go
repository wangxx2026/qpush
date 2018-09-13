package internalcmd

import (
	"encoding/json"
	"qpush/client"
	"qpush/pkg/logger"
	"qpush/server"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/go-kit/kit/metrics"
	"github.com/zhiqiangxu/qrpc"
)

// PushCmd do push
type PushCmd struct {
	pushCounterMetric metrics.Counter
}

// NewPushCmd returns a PushCmd instance
func NewPushCmd(pushCounterMetric metrics.Counter) *PushCmd {
	return &PushCmd{pushCounterMetric: pushCounterMetric}
}

// PushResp is resp for PushCmd
type PushResp struct {
	AppID    int
	OK       uint64
	NG       uint64
	MsgID    string
	NGDetail []string
	OKDetail []string
}

// ServeQRPC implements qrpc.Handler
func (cmd *PushCmd) ServeQRPC(writer qrpc.FrameWriter, frame *qrpc.RequestFrame) {

	logger.Debug("PushCmd called")

	var pushCmd client.PushCmd
	err := json.Unmarshal(frame.Payload, &pushCmd)
	if err != nil {
		logger.Error("failed to unmarshal PushCmd", string(frame.Payload))
		frame.Close()
		return
	}

	msg := client.Msg{
		MsgID: pushCmd.MsgID, Title: pushCmd.Title, Content: pushCmd.Content,
		Transmission: pushCmd.Transmission, Unfold: pushCmd.Unfold, PassThrough: pushCmd.PassThrough}
	bytes, err := json.Marshal(&msg)
	if err != nil {
		logger.Error("PushCmd failed to marshal client.Msg")
		frame.Close()
		return
	}

	var (
		count    uint64
		ngcount  uint64
		okngmu   sync.Mutex
		okDetail []string
		ngDetail []string
		wg       sync.WaitGroup
	)

	qserver := frame.ConnectionInfo().SC.Server()
	pushID := qserver.GetPushID()
	logger.Debug("PushCmd test1")
	// single mode
	if pushCmd.AppID != 0 && len(pushCmd.GUID) > 0 {
		ids := make([]string, 0, len(pushCmd.GUID))
		for _, guid := range pushCmd.GUID {
			ids = append(ids, server.GetAppGUID(pushCmd.AppID, guid))
		}

		logger.Debug("PushCmd ids", len(ids), ids)
		counterOKLabels := []string{"appid", strconv.Itoa(pushCmd.AppID), "kind", "pushok"}
		counterNGLabels := []string{"appid", strconv.Itoa(pushCmd.AppID), "kind", "pushng"}

		qserver.WalkConnByID(0, ids, func(w qrpc.FrameWriter, ci *qrpc.ConnectionInfo) {
			logger.Debug("PushCmd WalkConnByID")
			qrpc.GoFunc(&wg, func() {
				w.StartWrite(pushID, server.ForwardCmd, qrpc.PushFlag)
				w.WriteBytes(bytes)
				err := w.EndWrite()
				if err == nil {
					cmd.pushCounterMetric.With(counterOKLabels...).Add(1)
					logger.Info("PushCmd send ok", msg.MsgID, ci.GetID())
					atomic.AddUint64(&count, 1)

					okngmu.Lock()
					okDetail = append(okDetail, ci.GetID())
					okngmu.Unlock()
				} else {
					cmd.pushCounterMetric.With(counterNGLabels...).Add(1)
					logger.Info("PushCmd send ng", msg.MsgID, ci.GetID())
					atomic.AddUint64(&ngcount, 1)

					okngmu.Lock()
					ngDetail = append(ngDetail, ci.GetID())
					okngmu.Unlock()
				}
			})
		})

		wg.Wait()

		cmd.writeResp(writer, frame, &PushResp{AppID: pushCmd.AppID, OK: atomic.LoadUint64(&count), NG: atomic.LoadUint64(&ngcount), OKDetail: okDetail, NGDetail: ngDetail})
		return
	}

	// filter by city_ids
	cityIDMap := make(map[int]struct{})
	if len(pushCmd.CityIDS) > 0 {
		for _, cityID := range pushCmd.CityIDS {
			cityIDMap[cityID] = struct{}{}
		}
	}

	// filter by tag_ids
	tagIDMap := make(map[int]struct{})
	if len(pushCmd.TagIDS) > 0 {
		for _, tagID := range pushCmd.TagIDS {
			tagIDMap[tagID] = struct{}{}
		}
	}

	qserver.WalkConn(0, func(writer qrpc.FrameWriter, ci *qrpc.ConnectionInfo) bool {

		anything := ci.GetAnything()
		deviceInfo, ok := anything.(*server.DeviceInfo)
		if !ok { // anything is nil in this case
			return true
		}
		if pushCmd.AppID != deviceInfo.AppID {
			return true
		}
		// filter by os
		if pushCmd.OS != "" {
			if pushCmd.OS != deviceInfo.OS {
				return true
			}
		}
		// filter by city_id
		if len(cityIDMap) > 0 {

			_, ok := cityIDMap[deviceInfo.CityID]
			if !ok {
				return true
			}
		}
		// filter by tag_id
		if len(tagIDMap) > 0 {
			matched := false
			for _, tagID := range deviceInfo.TagIDS {
				_, ok := tagIDMap[tagID]
				if ok {
					matched = true
					break
				}
			}
			if !matched {
				return true
			}
		}

		qrpc.GoFunc(&wg, func() {
			writer.StartWrite(pushID, server.ForwardCmd, qrpc.PushFlag)
			writer.WriteBytes(bytes)
			err := writer.EndWrite()
			if err == nil {
				logger.Info("send ok", msg.MsgID, ci.GetID())
				atomic.AddUint64(&count, 1)

				okngmu.Lock()
				okDetail = append(okDetail, ci.GetID())
				okngmu.Unlock()
			} else {
				logger.Info("send ng", msg.MsgID, ci.GetID())
				atomic.AddUint64(&ngcount, 1)

				okngmu.Lock()
				ngDetail = append(ngDetail, ci.GetID())
				okngmu.Unlock()
			}
		})
		return true
	})
	wg.Wait()

	cmd.writeResp(writer, frame, &PushResp{AppID: pushCmd.AppID, OK: atomic.LoadUint64(&count), NG: atomic.LoadUint64(&ngcount), OKDetail: okDetail, NGDetail: ngDetail})
}

func (cmd *PushCmd) writeResp(writer qrpc.FrameWriter, frame *qrpc.RequestFrame, result *PushResp) {
	jsonwriter := server.JSONFrameWriter{FrameWriter: writer}

	jsonwriter.StartWrite(frame.RequestID, server.PushRespCmd, 0)
	jsonwriter.WriteJSON(result)
	err := jsonwriter.EndWrite()
	if err != nil {
		logger.Error("EndWrite fail", err)
	}

}
