package internalcmd

import (
	"encoding/json"
	"errors"
	"qpush/client"
	"qpush/modules/logger"
	"qpush/server"

	"github.com/zhiqiangxu/qrpc"
)

var (
	errUnmarshalFail = errors.New("Unmarshal failed")
)

// KillCmd do kill
type KillCmd struct {
}

// ServeQRPC implements qrpc.Handler
func (cmd *KillCmd) ServeQRPC(writer qrpc.FrameWriter, frame *qrpc.RequestFrame) {

	cmdInfo := &client.KillCmd{}
	err := json.Unmarshal(frame.Payload, cmdInfo)
	if err != nil {
		logger.Error(errUnmarshalFail, err)
		frame.Close()
		return
	}

	var ok bool
	ci := frame.ConnectionInfo()
	ci.SC.Server().WalkConnByID(0, []string{server.GetAppGUID(cmdInfo.AppID, cmdInfo.GUID)}, func(w qrpc.FrameWriter, ci *qrpc.ConnectionInfo) {
		ci.SC.Close()
		ok = true
	})

	cmd.writeResp(writer, frame, ok)

}

func (cmd *KillCmd) writeResp(writer qrpc.FrameWriter, frame *qrpc.RequestFrame, ok bool) {
	jsonwriter := server.JSONFrameWriter{FrameWriter: writer}

	jsonwriter.StartWrite(frame.RequestID, server.KillRespCmd, 0)
	jsonwriter.WriteJSON(ok)
	err := jsonwriter.EndWrite()
	if err != nil {
		logger.Error("EndWrite fail", err)
	}

}
