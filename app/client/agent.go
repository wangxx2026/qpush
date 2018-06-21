package main

import (
	"push-msg/client"
	"push-msg/client/impl"
	"push-msg/modules/logger"
	"push-msg/server"
)

func main() {
	agent := impl.NewAgent()
	conn := agent.Dial("localhost:8890")
	if conn == nil {
		logger.Error("failed to dial")
		return
	}

	logger.Debug("d1")
	pushCmd := &client.PushCmd{MsgID: 1, Message: "hello world"}
	ID, err := conn.SendCmd(server.PushCmd, pushCmd)
	if err != nil {
		logger.Error("SendCmd failed:", err)
	}

	logger.Debug("d2")
	cb := impl.NewCallBack(func(requestID uint64, cmd server.Cmd, bytes []byte) error {
		if ID == requestID {

		}
		logger.Info(requestID, cmd, string(bytes))
		return nil
	})
	logger.Debug("d3")
	err = conn.Subscribe(cb)
	logger.Debug("d4")
	logger.Error("Subscribe error", err)
}
