package main

import (
	"qpush/server"
	"qpush/server/cmd"
	"qpush/server/impl"
	"qpush/server/internalcmd"
)

func main() {
	serverHandler := &impl.ServerHandler{}
	serverHandler.RegisterCmd(server.LoginCmd, false, &cmd.LoginCmd{})
	serverHandler.RegisterCmd(server.AckCmd, false, cmd.NewAckCmd())

	serverHandler.RegisterCmd(server.PushCmd, true, &internalcmd.PushCmd{})

	serverConfig := server.Config{
		ReadBufferSize: server.DefaultReadBufferSize,
		Handler:        serverHandler}

	s := impl.NewServer(&serverConfig)
	s.ListenAndServe("localhost:8888", "localhost:8890")
}
