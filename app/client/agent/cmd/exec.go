package cmd

import (
	"fmt"
	"io"
	"os"
	"qpush/client"
	"qpush/modules/logger"
	"qpush/server"
	"qpush/server/internalcmd"
	"sync"

	"github.com/zhiqiangxu/qrpc"

	"github.com/kr/pty"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var execCmd = &cobra.Command{
	Use:   "exec [internal address] [cmd]",
	Short: "connect to [internal address] and exec [cmd]",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// config.Load("dev")
		internalAddress := args[0]
		cmdArg := args[1]

		conn, err := client.NewConnection(internalAddress, qrpc.ConnectionConfig{}, func(conn *client.Connection, frame *qrpc.Frame) {
			fmt.Println("pushed", string(frame.Payload))
		})
		if err != nil {
			logger.Error("NewConnection fail", err)
			return
		}

		cmdInfo := client.ExecCmd{Cmd: cmdArg}
		streamwriter, resp, err := conn.StreamRequest(server.ExecCmd, 0, cmdInfo)
		if err != nil {
			logger.Error("StreamRequest failed:", err)
			return
		}
		size, err := pty.GetsizeFull(os.Stdin)
		if err != nil {
			logger.Error("GetsizeFull failed:", err)
			return
		}
		streamwriter.StartWrite(internalcmd.SizeCmd)
		err = streamwriter.WriteBytes(size)
		if err != nil {
			logger.Error("WriteBytes failed:", err)
			return
		}
		err = streamwriter.EndWrite(false)
		if err != nil {
			logger.Error("EndWrite failed:", err)
			return
		}
		frame := resp.GetFrame()
		var wg sync.WaitGroup
		qrpc.GoFunc(&wg, func() {
			pipeInput(streamwriter)
		})
		nextFrame := frame
		for {
			if nextFrame.Flags&qrpc.StreamEndFlag == 0 {
				// fmt.Printf("called")
				os.Stdout.Write(nextFrame.Payload)
			} else {
				conn.Close(nil)
				fmt.Printf("Press any key to exit\r\n")
				break
			}

			nextFrame = <-frame.FrameCh()
			if nextFrame == nil {
				fmt.Println("nil NextFrame")
				break
			}
		}

		wg.Wait()

	}}

func init() {
	rootCmd.AddCommand(execCmd)
}

func pipeInput(streamwriter *client.StreamWriter) {

	// Set stdin in raw mode.
	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = terminal.Restore(int(os.Stdin.Fd()), oldState)
		// fmt.Println("terminal.Restored")
	}() // Best effort.

	_, err = io.Copy(&writer{streamwriter}, os.Stdin)
	if err != nil {
		// logger.Error("Copy error", err)
	}
}

type writer struct {
	streamwriter *client.StreamWriter
}

func (w *writer) Write(data []byte) (int, error) {

	w.streamwriter.StartWrite(internalcmd.InputCmd)
	w.streamwriter.StreamWriter.WriteBytes(data)
	err := w.streamwriter.EndWrite(false)
	if err != nil {
		return 0, err
	}

	return len(data), nil
}
