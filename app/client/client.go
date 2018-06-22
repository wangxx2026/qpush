package main

import (
	"qpush/client/impl"
	"qpush/modules/logger"
	"qpush/server"
	"time"
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
	go func() {
		for {
			time.Sleep(time.Second * 20)
			conn.SendCmd(server.HeartBeatCmd, nil)
		}
	}()
	err := conn.Subscribe(cb)
	logger.Error("Subscribe error", err)

}
