package cmd

import (
	"qpush/modules/logger"
	"qpush/server"

	"github.com/zhiqiangxu/qrpc"
)

// HeartBeatCmd do heartbeat
type HeartBeatCmd struct {
}

// ServeQRPC implements qrpc.Handler
func (cmd *HeartBeatCmd) ServeQRPC(writer qrpc.FrameWriter, frame *qrpc.RequestFrame) {
	logger.Debug("HeartBeatCmd called")
	writer.StartWrite(frame.RequestID, server.HeartBeatRespCmd, 0)
	err := writer.EndWrite()
	if err != nil {
		logger.Error("HeartBeatCmd", err)
	}

}
