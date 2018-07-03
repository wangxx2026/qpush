package agentcmd

import (
	"qpush/client"
	"qpush/client/impl"
	"qpush/modules/logger"
	"qpush/server"
	"strconv"

	"github.com/spf13/cobra"
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
		agent := impl.NewAgent()
		conn := agent.Dial(internalAddress)
		if conn == nil {
			logger.Error("failed to dial")
			return
		}

		cmdInfo := client.KillCmd{AppID: appid, GUID: guid}
		bytes, err := conn.SendCmdBlocking(server.KillCmd, &cmdInfo)
		if err != nil {
			logger.Error("SendCmdBlocking failed:", err)
		}

		logger.Info("result", string(bytes))
	}}

func init() {
	rootCmd.AddCommand(killCmd)
}
