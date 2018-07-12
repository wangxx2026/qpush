package agentcmd

import (
	"os"
	"qpush/client"
	"qpush/client/impl"
	"qpush/modules/logger"
	"qpush/server"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec [internal address] [cmd]",
	Short: "connect to [internal address] and exec [cmd]",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// config.Load("dev")
		internalAddress := args[0]
		cmdArg := args[1]
		agent := impl.NewAgent()
		conn := agent.Dial(internalAddress)
		if conn == nil {
			logger.Error("failed to dial")
			return
		}

		cmdInfo := client.ExecCmd{Cmd: cmdArg}
		requestID, err := conn.SendCmd(server.ExecCmd, &cmdInfo)
		if err != nil {
			logger.Error("SendCmdBlocking failed:", err)
		}
		conn.Subscribe(impl.NewCallBack(func(ID uint64, cmd server.Cmd, bytes []byte) bool {
			if cmd == server.ExecRespCmd {
				os.Stdout.Write(bytes)
			}
			if requestID == ID {
				return false
			}
			return true
		}))

	}}

func init() {
	rootCmd.AddCommand(execCmd)
}
