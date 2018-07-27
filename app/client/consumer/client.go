package main

import (
	"fmt"
	"qpush/client"
	"qpush/modules/logger"
	"qpush/server"
	"strconv"
	"time"

	"github.com/zhiqiangxu/qrpc"

	"github.com/spf13/cobra"
)

func main() {

	var rootCmd = &cobra.Command{
		Use:   "qpushclient [public address] [appid] [appkey] [guid]",
		Short: "login into [public address] as [guid] and subscribe for messages",
		Args:  cobra.MinimumNArgs(4),
		Run: func(cmd *cobra.Command, args []string) {

			appID, err := strconv.Atoi(args[1])
			if err != nil {
				logger.Error("invalid appid")
				return
			}
			loginCmd := client.LoginCmd{GUID: args[3], AppID: appID, AppKey: args[2]}

			c, err := qrpc.NewConnection(args[0], qrpc.ConnectionConfig{}, func(conn *qrpc.Connection, frame *qrpc.Frame) {
				fmt.Printf("%s\n", frame.Payload)
			})
			if err != nil {
				panic(err)
			}
			conn := &client.Connection{Connection: c}
			_, resp, err := conn.Request(server.LoginCmd, 0, loginCmd)
			if err != nil {
				panic(err)
			}
			if frame := resp.GetFrame(); frame != nil {
				fmt.Printf("login resp:%s\n", frame.Payload)
			} else {
				panic("login failed")
			}

			go func() {
				for {
					time.Sleep(time.Second * 300)
					_, resp, err := conn.Connection.Request(server.HeartBeatCmd, 0, nil)
					if err != nil {
						panic(err)
					}
					fmt.Printf("hearbeat resp:%v\n", resp.GetFrame().Payload)
				}
			}()
			conn.Wait()

		}}

	rootCmd.Execute()

}
