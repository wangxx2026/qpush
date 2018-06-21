package main

import (
	"push-msg/client/impl"
	"push-msg/modules/logger"
	"push-msg/server"
)

func main() {
	client := impl.NewClient()
	conn := client.Dial("localhost:8888", "guid")
	if conn == nil {
		logger.Error("failed to dial")
	}

	cb := impl.NewCallBack(func(requestID uint64, cmd server.Cmd, bytes []byte) error {
		logger.Info(requestID, cmd, string(bytes))
		return nil
	})
	err := conn.Subscribe(cb)
	logger.Error("Subscribe error", err)

}
