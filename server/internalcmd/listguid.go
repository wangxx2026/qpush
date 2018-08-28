package internalcmd

import (
	"qpush/pkg/logger"
	"qpush/server"
	"sort"

	"github.com/zhiqiangxu/qrpc"
)

// ListGUIDCmd lists guids
type ListGUIDCmd struct {
}

// ServeQRPC implements qrpc.Handler
func (cmd *ListGUIDCmd) ServeQRPC(writer qrpc.FrameWriter, frame *qrpc.RequestFrame) {

	logger.Debug("ListGUIDCmd called")
	qserver := frame.ConnectionInfo().SC.Server()
	var result server.DeviceInfoSlice
	qserver.WalkConn(0, func(w qrpc.FrameWriter, ci *qrpc.ConnectionInfo) bool {
		deviceInfo, ok := ci.GetAnything().(*server.DeviceInfo)
		if ok {
			result = append(result, deviceInfo)
		}

		return true
	})

	sort.Sort(result)

	cmd.writeResp(writer, frame, result)

}

func (cmd *ListGUIDCmd) writeResp(writer qrpc.FrameWriter, frame *qrpc.RequestFrame, result interface{}) {
	jsonwriter := server.JSONFrameWriter{FrameWriter: writer}

	jsonwriter.StartWrite(frame.RequestID, server.ListGUIDRespCmd, 0)
	jsonwriter.WriteJSON(result)
	err := jsonwriter.EndWrite()
	if err != nil {
		logger.Error("EndWrite fail", err)
	}

}
