package main

import (
	"qpush/client/impl"
	"qpush/modules/logger"
	"qpush/server"
)

func main() {
	client := impl.NewClient()
	conn := client.Dial("localhost:8888", "guid")
	if conn == nil {
		logger.Error("failed to dial")
	}

	cb := impl.NewCallBack(func(requestID uint64, cmd server.Cmd, bytes []byte) bool {
		logger.Info(requestID, cmd, string(bytes))
		return true
	})
	err := conn.Subscribe(cb)
	logger.Error("Subscribe error", err)

}
