package main

import (
	"qpush/client"
	"qpush/client/impl"
	"qpush/modules/logger"
	"qpush/server"
	"strconv"

	"github.com/spf13/cobra"
)

func main() {

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
			ID, err := conn.SendCmd(server.PushCmd, pushCmd)
			if err != nil {
				logger.Error("SendCmd failed:", err)
			}

			cb := impl.NewCallBack(func(requestID uint64, cmd server.Cmd, bytes []byte) bool {

				if ID == requestID {
					logger.Info("got reply")
					logger.Info(requestID, cmd, string(bytes))
					return false
				}

				return true
			})

			err = conn.Subscribe(cb)

			if err != nil {
				logger.Error("Subscribe error", err)
			}
		}}
	var rootCmd = &cobra.Command{
		Use: "qpushagent",
	}
	rootCmd.AddCommand(pushCmd)
	rootCmd.Execute()
}
