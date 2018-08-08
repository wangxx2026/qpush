package cmd

import (
	"qpush/pkg/logger"
	"qpush/server"

	"github.com/spf13/cobra"
	"github.com/zhiqiangxu/qrpc"
)

var listGUIDCmd = &cobra.Command{
	Use:   "list [internal address]",
	Short: "connect to [internal address] and list guid",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		internalAddress := args[0]
		conn, err := qrpc.NewConnection(internalAddress, qrpc.ConnectionConfig{}, nil)
		if err != nil {
			panic(err)
		}

		_, resp, err := conn.Request(server.ListGUIDCmd, 0, nil)
		frame, err := resp.GetFrame()
		if err != nil {
			panic(err)
		}

		logger.Info("result", string(frame.Payload))
	}}

func init() {
	rootCmd.AddCommand(listGUIDCmd)
}
