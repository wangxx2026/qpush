package main

import (
	"qpush/client"
	"qpush/client/impl"
	"qpush/modules/logger"
	"qpush/server"
)

func main() {
	agent := impl.NewAgent()
	conn := agent.Dial("localhost:8890")
	if conn == nil {
		logger.Error("failed to dial")
		return
	}

	logger.Debug("d1")
	pushCmd := &client.PushCmd{MsgID: 1, Title: "hello title", Content: "hello content"}
	ID, err := conn.SendCmd(server.PushCmd, pushCmd)
	if err != nil {
		logger.Error("SendCmd failed:", err)
	}

	logger.Debug("d2")
	cb := impl.NewCallBack(func(requestID uint64, cmd server.Cmd, bytes []byte) bool {
		if ID == requestID {

		}
		logger.Info(requestID, cmd, string(bytes))
		return true
	})
	logger.Debug("d3")
	err = conn.Subscribe(cb)
	logger.Debug("d4")
	logger.Error("Subscribe error", err)
}
