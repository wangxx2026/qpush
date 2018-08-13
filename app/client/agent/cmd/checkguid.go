package cmd

import (
	"context"
	"qpush/client"
	"qpush/pkg/logger"
	"qpush/server"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/zhiqiangxu/qrpc"
)

var checkGUIDCmd = &cobra.Command{
	Use:   "checkguid [internal address] [appid] [guid]",
	Short: "connect to [internal address] and check specific connection",
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		internalAddress := args[0]
		appid, err := strconv.Atoi(args[1])
		if err != nil {
			logger.Error("invalid appid")
			return
		}
		guid := args[2]

		cmdInfo := client.CheckGUIDCmd{AppID: appid, GUID: guid}
		api := client.NewAPI([]string{internalAddress}, qrpc.ConnectionConfig{}, nil)
		frame, err := api.CallForFrame(context.Background(), server.CheckGUIDCmd, cmdInfo)
		if err != nil {
			panic(err)
		}

		logger.Info("result", string(frame.Payload))
	}}

func init() {
	rootCmd.AddCommand(checkGUIDCmd)
}
