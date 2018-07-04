package internalcmd

import (
	"net"
	"qpush/server"
)

// ListGUIDCmd lists guids
type ListGUIDCmd struct {
}

type guid struct {
	GUID  string `json:"guid"`
	AppID int    `json:"app_id"`
}

// Call implements CmdHandler
func (cmd *ListGUIDCmd) Call(param *server.CmdParam) (server.Cmd, interface{}, error) {

	var result []guid
	param.Server.Walk(func(conn net.Conn, ctx *server.ConnectionCtx) bool {
		g := guid{GUID: ctx.GUID, AppID: ctx.AppID}
		result = append(result, g)

		return true
	})
	return server.ListGUIDRespCmd, result, nil

}
