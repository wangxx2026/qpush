package main

import (
	"push-msg/conf"
	"push-msg/server"
	"push-msg/server/cmd"
)

func main() {
	conf.DefaultServerHandler.RegisterCmd("login", false, &cmd.LoginCmd{})

	serverConfig := conf.ServerConfig{
		ReadBufferSize: conf.DefaultReadBufferSize,
		Handler:        conf.DefaultServerHandler}

	s := server.New(&serverConfig)
	s.ListenAndServe("localhost:8888", "localhost:8890")
}
