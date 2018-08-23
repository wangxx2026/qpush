package cmd

import (
	"qpush/pkg/logger"
	"qpush/server"
	"unsafe"

	"github.com/zhiqiangxu/qrpc"
)

// HeartBeatCmd do heartbeat
type HeartBeatCmd struct {
}

// ServeQRPC implements qrpc.Handler
func (cmd *HeartBeatCmd) ServeQRPC(writer qrpc.FrameWriter, frame *qrpc.RequestFrame) {
	ci := frame.ConnectionInfo()
	serveconn := ci.SC
	mem := unsafe.Pointer(serveconn)

	logger.Debug(mem, "HeartBeatCmd called")
	writer.StartWrite(frame.RequestID, server.HeartBeatRespCmd, 0)
	err := writer.EndWrite()
	if err != nil {
		logger.Error(mem, "HeartBeatCmd", err)
	}

}
