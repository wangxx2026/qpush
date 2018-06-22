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

			ID, err := conn.SendCmd(server.StatusCmd, nil)
			if err != nil {
				logger.Error("failed to send status cmd", err)
				return
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
			ID, err := conn.SendCmd(server.KillCmd, &cmdInfo)
			if err != nil {
				logger.Error("failed to send kill cmd", err)
				return
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
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(killCmd)
	rootCmd.Execute()
}
