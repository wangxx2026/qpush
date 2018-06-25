package agentcmd

import (
	"qpush/client/impl"
	"qpush/modules/logger"
	"qpush/server"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [internal address]",
	Short: "connect to [internal address] and query server status",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		internalAddress := args[0]
		agent := impl.NewAgent()
		conn := agent.Dial(internalAddress)
		if conn == nil {
			logger.Error("failed to dial")
			return
		}

		bytes, err := conn.SendCmdBlocking(server.StatusCmd, nil)
		if err != nil {
			logger.Error("SendCmdBlocking failed:", err)
		}

		logger.Info("result", string(bytes))
	}}

func init() {
	rootCmd.AddCommand(statusCmd)
}
