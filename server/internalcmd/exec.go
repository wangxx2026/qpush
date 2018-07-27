package internalcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"qpush/client"
	"qpush/modules/logger"
	"qpush/modules/util"
	"qpush/server"
	"sync"

	"github.com/zhiqiangxu/qrpc"

	"github.com/kr/pty"
)

const (
	// SizeCmd for set size
	SizeCmd qrpc.Cmd = iota
	// InputCmd for pipe stdin
	InputCmd
)

// ExecCmd do exec
type ExecCmd struct {
}

// ServeQRPC implements qrpc.Handler
func (cmd *ExecCmd) ServeQRPC(writer qrpc.FrameWriter, frame *qrpc.RequestFrame) {
	var execCmd client.ExecCmd
	err := json.Unmarshal(frame.Payload, &execCmd)
	if err != nil {
		logger.Error(server.ErrUnMarshalFail)
		frame.Close()
		return
	}

	err = cmd.runCmd(writer, frame, execCmd.Cmd)
	logger.Info("runCmd", err)

	jsonwriter := server.JSONFrameWriter{FrameWriter: writer}
	jsonwriter.StartWrite(frame.RequestID, server.ExecRespCmd, qrpc.StreamFlag|qrpc.StreamEndFlag|qrpc.NBFlag)
	jsonwriter.WriteJSON(client.ExecRespCmd{Err: util.ToString(err)})
	err = jsonwriter.EndWrite()
	if err != nil {
		logger.Error("ExecCmd EndWrite fail", err)
	}
}

func (cmd *ExecCmd) runCmd(writer qrpc.FrameWriter, frame *qrpc.RequestFrame, bashCmd string) (errResp error) {

	if bashCmd == "" {
		return server.ErrInvalidParam
	}
	logger.Debug("runCmd", frame.RequestID, frame.Flags)
	// Create arbitrary command.
	c := exec.Command("bash", "-c", bashCmd)

	// Start the command with a pty.
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}

	// Make sure to close the pty at the end.
	defer func() {
		_ = ptmx.Close()

		errResp = c.Wait()
	}() // Best effort.

	funcDone := make(chan bool)
	ctx := frame.Context()
	// kill process if canceled
	var wg sync.WaitGroup
	qrpc.GoFunc(&wg, func() {
		select {
		case <-ctx.Done():
			logger.Info("ctx done")
			err = c.Process.Kill()
			if err != nil {
				logger.Error("Kill failed", err)
			}
		case <-funcDone:
		}
	})

	qrpc.GoFunc(&wg, func() {
		for {
			select {
			case nextFrame := <-frame.FrameCh():
				if nextFrame == nil {
					logger.Debug("nextFrame is nil")
					return
				}
				switch nextFrame.Cmd {
				case SizeCmd:
					var size pty.Winsize
					err := json.Unmarshal(nextFrame.Payload, &size)
					if err != nil {
						logger.Error("Unmarshal", err)
						continue
					}

					err = pty.Setsize(ptmx, &size)
					if err != nil {
						logger.Error("SetSize", err)
						continue
					}
				case InputCmd:
					fmt.Printf("%s", nextFrame.Payload)
					ptmx.Write(nextFrame.Payload)
				}
			case <-funcDone:
				return
			}

		}

	})

	_, err = io.Copy(&connWriter{qrpc.NewStreamWriter(writer, frame.RequestID, frame.Flags), ctx}, ptmx)

	close(funcDone)
	wg.Wait()

	return
}

type connWriter struct {
	writer qrpc.StreamWriter
	ctx    context.Context
}

func (w *connWriter) Write(data []byte) (int, error) {

	w.writer.StartWrite(server.ExecRespCmd)
	w.writer.WriteBytes(data)
	err := w.writer.EndWrite(false)
	if err != nil {
		logger.Error("EndWrite", err)
	}

	return len(data), err
}
