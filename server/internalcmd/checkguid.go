package internalcmd

import (
	"encoding/json"
	"qpush/client"
	"qpush/pkg/logger"
	"qpush/server"

	"github.com/zhiqiangxu/qrpc"
)

// CheckGUIDCmd check guid online state
type CheckGUIDCmd struct {
}

// ServeQRPC implements qrpc.Handler
func (cmd *CheckGUIDCmd) ServeQRPC(writer qrpc.FrameWriter, frame *qrpc.RequestFrame) {

	var checkCmd client.CheckGUIDCmd
	err := json.Unmarshal(frame.Payload, &checkCmd)
	if err != nil {
		logger.Error(server.ErrUnMarshalFail)
		frame.Close()
		return
	}

	logger.Debug("CheckGUIDCmd called")
	qserver := frame.ConnectionInfo().SC.Server()
	var result interface{}
	qserver.WalkConnByID(0, []string{server.GetAppGUID(checkCmd.AppID, checkCmd.GUID)}, func(w qrpc.FrameWriter, ci *qrpc.ConnectionInfo) {
		if ci.Anything != nil {
			result = ci.Anything
		}
	})

	cmd.writeResp(writer, frame, result)

}

func (cmd *CheckGUIDCmd) writeResp(writer qrpc.FrameWriter, frame *qrpc.RequestFrame, result interface{}) {
	jsonwriter := server.JSONFrameWriter{FrameWriter: writer}

	jsonwriter.StartWrite(frame.RequestID, server.CheckGUIDRespCmd, 0)
	jsonwriter.WriteJSON(result)
	err := jsonwriter.EndWrite()
	if err != nil {
		logger.Error("EndWrite fail", err)
	}

}
