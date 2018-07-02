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
	if pushCmd.AppID != 0 && pushCmd.GUID != "" {
		return server.PushRespCmd, true, s.SendTo(pushCmd.AppID, pushCmd.GUID, packet)
	}

	s.Walk(func(conn net.Conn, writeChan chan []byte) bool {
		if selfConn != conn {
			ctx := s.GetCtx(conn)
			if ctx.Internal {
				return true
			}
			select {
			case writeChan <- packet:
			default:
				logger.Error("writeChan blocked for", ctx.GUID)
			}
		}

		return true
	})

	return server.PushRespCmd, true, nil
}
