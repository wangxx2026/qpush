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
			bytes, err := conn.SendCmdBlocking(server.PushCmd, pushCmd)
			if err != nil {
				logger.Error("SendCmdBlocking failed:", err)
			}

			logger.Info("result", string(bytes))
		}}

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
	var killCmd = &cobra.Command{
		Use:   "kill [internal address] [guid]",
		Short: "connect to [internal address] and kill specified connection",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			internalAddress := args[0]
			guid := args[1]
			agent := impl.NewAgent()
			conn := agent.Dial(internalAddress)
			if conn == nil {
				logger.Error("failed to dial")
				return
			}

			cmdInfo := client.KillCmd{GUID: guid}
			bytes, err := conn.SendCmdBlocking(server.KillCmd, &cmdInfo)
			if err != nil {
				logger.Error("SendCmdBlocking failed:", err)
			}

			logger.Info("result", string(bytes))
		}}
	var rootCmd = &cobra.Command{
		Use: "qpushagent",
	}
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(killCmd)
	rootCmd.Execute()
}
