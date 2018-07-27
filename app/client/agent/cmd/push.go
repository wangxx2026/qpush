package cmd

import (
	"fmt"
	"qpush/client"
	"qpush/modules/logger"
	"qpush/server"

	"github.com/spf13/cobra"
	"github.com/zhiqiangxu/qrpc"
)

var pushCmd = &cobra.Command{
	Use:   "push [internal address] [msg id] [title] [content]",
	Short: "connect to [internal address] and send push cmd",
	Args:  cobra.MinimumNArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		internalAddress := args[0]
		msgID := args[1]
		title := args[2]
		content := args[3]

		conn, err := client.NewConnection(internalAddress, qrpc.ConnectionConfig{}, func(conn *client.Connection, frame *qrpc.Frame) {
			fmt.Println("pushed", string(frame.Payload))
		})
		if err != nil {
			logger.Error("NewConnection fail", err)
			return
		}

		pushCmd := client.PushCmd{Msg: client.Msg{MsgID: msgID, Title: title, Content: content}}
		_, resp, err := conn.Request(server.PushCmd, 0, pushCmd)
		if err != nil {
			logger.Error("Request failed:", err)
			return
		}
		frame := resp.GetFrame()
		if frame == nil {
			logger.Error("GetFrame failed: nil")
			return
		}

		logger.Info("result", string(frame.Payload))
	}}

func init() {
	rootCmd.AddCommand(pushCmd)
}
