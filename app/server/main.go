package main

import (
	"qpush/modules/logger"
	"qpush/server"
	"qpush/server/cmd"
	"qpush/server/impl"
	"qpush/server/internalcmd"
	"time"
)

const (
	// ServerHeartBeatInteval is interval for server heartbeat
	ServerHeartBeatInteval = time.Second * 3
	// PublicAddress is for public
	PublicAddress = "localhost:8888"
	// InternalAddress is for internal
	InternalAddress = "localhost:8890"
)

func main() {

	serverHandler := &impl.ServerHandler{}
	serverHandler.RegisterCmd(server.LoginCmd, false, &cmd.LoginCmd{})
	serverHandler.RegisterCmd(server.AckCmd, false, cmd.NewAckCmd())
	serverHandler.RegisterCmd(server.HeartBeatCmd, false, &cmd.HeartBeatCmd{})

	serverHandler.RegisterCmd(server.PushCmd, true, &internalcmd.PushCmd{})

	hbConfig := server.HeartBeatConfig{
		Callback: func() error {
			logger.Info("heartbeat called")
			//TODO call interface
			return nil
		},
		Interval: ServerHeartBeatInteval}
	serverConfig := server.Config{
		ReadBufferSize: server.DefaultReadBufferSize,
		Handler:        serverHandler,
		HBConfig:       hbConfig}

	s := impl.NewServer(&serverConfig)
	s.ListenAndServe(PublicAddress, InternalAddress)

}
