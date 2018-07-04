package internalcmd

import (
	"encoding/json"
	"errors"
	"net"
	"qpush/client"
	"qpush/modules/logger"
	"qpush/server"
	cmdpackage "qpush/server/cmd"
	"qpush/server/impl"
)

var (
	errMarshalFail   = errors.New("failed to marshal")
	errUnMarshalFail = errors.New("failed to unmarshal")
)

// PushCmd do push
type PushCmd struct {
}

// Call implements CmdHandler
func (cmd *PushCmd) Call(param *server.CmdParam) (server.Cmd, interface{}, error) {
	s := param.Server
	selfConn := param.Conn
	message := param.Param

	var pushCmd client.PushCmd
	err := json.Unmarshal(message, &pushCmd)
	if err != nil {
		logger.Error("failed to unmarshal")
		return server.ErrorCmd, nil, errUnMarshalFail
	}

	msg := cmdpackage.Msg{
		MsgID: pushCmd.MsgID, Title: pushCmd.Title, Content: pushCmd.Content,
		Transmission: pushCmd.Transmission, Unfold: pushCmd.Unfold}
	bytes, err := json.Marshal(&msg)
	if err != nil {
		logger.Error("failed to marshal")
		return server.ErrorCmd, false, errMarshalFail
	}
	packet := impl.MakePacket(param.RequestID, server.ForwardCmd, bytes)

	// single mode
	if pushCmd.AppID != 0 && len(pushCmd.GUID) > 0 {
		return server.PushRespCmd, s.SendTo(pushCmd.AppID, pushCmd.GUID, packet), nil
	}

	var count int
	s.Walk(func(conn net.Conn, ctx *server.ConnectionCtx) bool {
		if selfConn != conn {
			if ctx.Internal {
				return true
			}
			select {
			case ctx.WriteChan <- packet:
				count++
			default:
				logger.Error("writeChan blocked for", ctx.GUID)
			}
		}

		return true
	})

	return server.PushRespCmd, count, nil
}

// Status returns status of this cmd
func (cmd *PushCmd) Status() interface{} {
	return nil
}
