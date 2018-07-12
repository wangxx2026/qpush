package internalcmd

import (
	"encoding/json"
	"io"
	"os/exec"
	"qpush/client"
	"qpush/modules/logger"
	"qpush/modules/util"
	"qpush/server"

	"github.com/kr/pty"
)

// ExecCmd do exec
type ExecCmd struct {
}

// Call implements CmdHandler
func (cmd *ExecCmd) Call(param *server.CmdParam) (server.Cmd, interface{}, error) {
	var execCmd client.ExecCmd
	err := json.Unmarshal(param.Param, &execCmd)

	if err != nil {
		logger.Error(server.ErrUnMarshalFail)
		return server.ErrorCmd, nil, server.ErrUnMarshalFail
	}

	err = cmd.runCmd(param, execCmd.Cmd)

	return server.ExecRespCmd, &client.ExecRespCmd{Err: util.ToString(err)}, nil
}

func (cmd *ExecCmd) runCmd(param *server.CmdParam, bashCmd string) error {

	if bashCmd == "" {
		return server.ErrInvalidParam
	}
	// Create arbitrary command.
	c := exec.Command("bash", "-c", bashCmd)

	// Start the command with a pty.
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}

	// Make sure to close the pty at the end.
	defer func() { _ = ptmx.Close() }() // Best effort.

	funcDone := make(chan bool)
	defer func() { close(funcDone) }()
	// kill process if socket closed
	go func() {
		select {
		case <-param.Ctx.CloseChan:
			c.Process.Kill()
		case <-funcDone:
		}

	}()

	_, err = io.Copy(&connWriter{param}, ptmx)

	if err != nil {
		return err
	}

	return c.Wait()
}

type connWriter struct {
	param *server.CmdParam
}

func (w *connWriter) Write(data []byte) (int, error) {

	// logger.Debug("Write called")
	copyData := make([]byte, len(data))
	copy(copyData, data)

	select {
	case <-w.param.Ctx.CloseChan:
		return 0, server.ErrConnectionClosed
	default:
	}

	select {
	case w.param.Ctx.WriteChan <- server.MakePacket(0, server.ExecRespCmd, copyData):
		// logger.Debug(len(copyData), string(copyData))
		return len(data), nil
	case <-w.param.Ctx.CloseChan:
		return 0, server.ErrConnectionClosed
	}

}
