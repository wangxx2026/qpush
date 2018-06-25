package agentcmd

import (
	"qpush/client/impl"
	"qpush/modules/logger"
	"qpush/server"

	"github.com/spf13/cobra"
)

var killAllCmd = &cobra.Command{
	Use:   "killall [internal address]",
	Short: "connect to [internal address] and kill all connection",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		internalAddress := args[0]
		agent := impl.NewAgent()
		conn := agent.Dial(internalAddress)
		if conn == nil {
			logger.Error("failed to dial")
			return
		}

		bytes, err := conn.SendCmdBlocking(server.KillAllCmd, nil)
		if err != nil {
			logger.Error("SendCmdBlocking failed:", err)
		}

		logger.Info("result", string(bytes))
	}}

func init() {
	rootCmd.AddCommand(killAllCmd)
}
