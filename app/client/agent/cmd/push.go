package cmd

import (
	"context"
	"qpush/client"
	"qpush/pkg/logger"
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

		api := client.NewAPI([]string{internalAddress}, qrpc.ConnectionConfig{}, nil)
		pushCmd := client.PushCmd{Msg: client.Msg{MsgID: msgID, Title: title, Content: content}}
		frame, err := api.CallForFrame(context.Background(), server.PushCmd, pushCmd)
		if err != nil {
			panic(err)
		}

		logger.Info("result", string(frame.Payload))
	}}

func init() {
	rootCmd.AddCommand(pushCmd)
}
