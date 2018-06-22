package main

import (
	"qpush/client/impl"
	"qpush/modules/logger"
	"qpush/server"
	"time"

	"github.com/spf13/cobra"
)

func main() {

	var rootCmd = &cobra.Command{
		Use:   "qpushclient [public address] [guid]",
		Short: "login into [public address] as [guid] and subscribe for messages",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			client := impl.NewClient()
			conn := client.Dial(args[0], args[1])
			if conn == nil {
				logger.Error("failed to dial")
				return
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
		}}
	rootCmd.Execute()

}
