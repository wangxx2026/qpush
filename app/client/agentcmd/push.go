package agentcmd

import (
	"qpush/client"
	"qpush/client/impl"
	"qpush/modules/logger"
	"qpush/server"
	"strconv"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [internal address] [msg id] [title] [content]",
	Short: "connect to [internal address] and send push cmd",
	Args:  cobra.MinimumNArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		internalAddress := args[0]
		msgID, err := strconv.Atoi(args[1])
		if err != nil {
			logger.Error("Atoi failed:", err)
			return
		}
		title := args[2]
		content := args[3]

		agent := impl.NewAgent()
		conn := agent.Dial(internalAddress)
		if conn == nil {
			logger.Error("failed to dial")
			return
		}

		pushCmd := &client.PushCmd{MsgID: msgID, Title: title, Content: content}
		bytes, err := conn.SendCmdBlocking(server.PushCmd, pushCmd)
		if err != nil {
			logger.Error("SendCmdBlocking failed:", err)
		}

		logger.Info("result", string(bytes))
	}}

func init() {
	rootCmd.AddCommand(pushCmd)
}
