package cmd

import (
	"qpush/client"
	"qpush/modules/logger"
	"qpush/server"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/zhiqiangxu/qrpc"
)

var killCmd = &cobra.Command{
	Use:   "kill [internal address] [appid] [guid]",
	Short: "connect to [internal address] and kill specified connection",
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		internalAddress := args[0]
		appid, err := strconv.Atoi(args[1])
		if err != nil {
			logger.Error("invalid appid")
			return
		}
		guid := args[2]
		conn, err := client.NewConnection(internalAddress, qrpc.ConnectionConfig{}, nil)
		if err != nil {
			panic(err)
		}

		cmdInfo := client.KillCmd{AppID: appid, GUID: guid}
		_, resp, err := conn.Request(server.KillCmd, 0, cmdInfo)
		if err != nil {
			panic(err)
		}

		frame := resp.GetFrame()

		logger.Info("result", string(frame.Payload))
	}}

func init() {
	rootCmd.AddCommand(killCmd)
}
