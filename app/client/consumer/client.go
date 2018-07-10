package main

import (
	"qpush/client"
	"qpush/client/impl"
	"qpush/modules/logger"
	"qpush/server"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

func main() {

	var rootCmd = &cobra.Command{
		Use:   "qpushclient [public address] [appid] [appkey] [guid]",
		Short: "login into [public address] as [guid] and subscribe for messages",
		Args:  cobra.MinimumNArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			cli := impl.NewClient()
			appID, err := strconv.Atoi(args[1])
			if err != nil {
				logger.Error("invalid appid")
				return
			}
			loginCmd := client.LoginCmd{GUID: args[3], AppID: appID, AppKey: args[2]}
			conn := cli.Dial(args[0], &loginCmd)
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
					time.Sleep(time.Second * 60)
					conn.SendCmd(server.HeartBeatCmd, nil)
				}
			}()
			err = conn.Subscribe(cb)
			logger.Error("Subscribe error", err)
		}}

	rootCmd.Execute()

}
