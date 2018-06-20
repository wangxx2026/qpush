package main

import (
	"push-msg/server"
	"push-msg/server/cmd"
	"push-msg/server/impl"
)

func main() {
	serverHandler := &impl.ServerHandler{}
	serverHandler.RegisterCmd("login", false, &cmd.LoginCmd{})

	serverConfig := server.Config{
		ReadBufferSize: server.DefaultReadBufferSize,
		Handler:        serverHandler}

	s := impl.NewServer(&serverConfig)
	s.ListenAndServe("localhost:8888", "localhost:8890")
}
