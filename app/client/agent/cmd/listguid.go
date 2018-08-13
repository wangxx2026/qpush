package cmd

import (
	"context"
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

		api := qrpc.NewAPI([]string{internalAddress}, qrpc.ConnectionConfig{}, nil)
		frame, err := api.Call(context.Background(), server.ListGUIDCmd, nil)
		if err != nil {
			panic(err)
		}

		logger.Info("result", string(frame.Payload))
	}}

func init() {
	rootCmd.AddCommand(listGUIDCmd)
}
