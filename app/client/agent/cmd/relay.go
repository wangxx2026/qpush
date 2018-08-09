package cmd

import (
	"bufio"
	"fmt"
	"os"
	"qpush/pkg/logger"
	"qpush/server"

	"qpush/server/internalcmd"

	"github.com/spf13/cobra"
	"github.com/zhiqiangxu/qrpc"
)

var relayCmd = &cobra.Command{
	Use:   "relay [internal address]",
	Short: "connect to [internal address] and start random relay chat",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		internalAddress := args[0]

		conn, err := qrpc.NewConnection(internalAddress, qrpc.ConnectionConfig{}, nil)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		streamwriter, response, err := conn.StreamRequest(server.RelayCmd, qrpc.StreamFlag, nil)
		if err != nil {
			panic(err)
		}

		fmt.Println("waiting for peer..")
		frame, err := response.GetFrame()
		if err != nil {
			panic(err)
		}

		if frame.Cmd != internalcmd.SessionStartCmd {
			panic("expecting SessionStartCmd")
		}
		fmt.Println("peer matched!")

		go func() {
			for {
				msgFrame := <-frame.FrameCh()
				if msgFrame == nil {
					logger.Debug("nil frame")
					os.Exit(0)
				}
				if msgFrame.Cmd == internalcmd.MsgCmd {
					fmt.Printf("\x1b[32m%s\x1b[0m> \n", string(msgFrame.Payload))
				}
			}
		}()
		reader := bufio.NewReader(os.Stdin)
		for {
			text, _ := reader.ReadString('\n')
			streamwriter.StartWrite(internalcmd.MsgCmd)
			streamwriter.WriteBytes([]byte(text))
			err = streamwriter.EndWrite(false)
			if err != nil {
				fmt.Println("session finished!")
				return
			}
		}

	}}

func init() {
	rootCmd.AddCommand(relayCmd)
}
