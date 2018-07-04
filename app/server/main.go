package main

import (
	"fmt"
	_ "net/http/pprof"
	"qpush/modules/config"
	"qpush/server"
	"qpush/server/cmd"
	"qpush/server/impl"
	"qpush/server/internalcmd"
	"time"

	"github.com/spf13/cobra"
)

const (
	// ServerHeartBeatInteval is interval for server heartbeat
	ServerHeartBeatInteval = time.Second * 30
	// DefaultPublicAddress is for public
	DefaultPublicAddress = "localhost:8888"
	// DefaultInternalAddress is for internal
	DefaultInternalAddress = "localhost:8890"
)

var (
	env string
)

func main() {

	var rootCmd = &cobra.Command{
		Use:   "qpushserver [public address] [internal address]",
		Short: "listen and server at specified address",
		Run: func(cobraCmd *cobra.Command, args []string) {
			var (
				publicAddress   string
				internalAddress string
			)
			if len(args) > 0 {
				publicAddress = args[0]
			} else {
				publicAddress = DefaultPublicAddress
			}

			if len(args) > 1 {
				internalAddress = args[1]
			} else {
				internalAddress = DefaultInternalAddress
			}

			serverHandler := &impl.ServerHandler{}
			serverHandler.RegisterCmd(server.LoginCmd, false, &cmd.LoginCmd{})
			serverHandler.RegisterCmd(server.AckCmd, false, cmd.NewAckCmd())
			serverHandler.RegisterCmd(server.HeartBeatCmd, false, &cmd.HeartBeatCmd{})

			serverHandler.RegisterCmd(server.PushCmd, true, &internalcmd.PushCmd{})
			serverHandler.RegisterCmd(server.StatusCmd, true, &internalcmd.StatusCmd{})
			serverHandler.RegisterCmd(server.KillCmd, true, &internalcmd.KillCmd{})
			serverHandler.RegisterCmd(server.KillAllCmd, true, &internalcmd.KillAllCmd{})
			serverHandler.RegisterCmd(server.ListGUIDCmd, true, &internalcmd.ListGUIDCmd{})

			hbConfig := server.HeartBeatConfig{
				Callback: func() error {
					// logger.Info("heartbeat called")
					//TODO call interface
					return nil
				},
				Interval: ServerHeartBeatInteval}
			serverConfig := server.Config{
				ReadBufferSize: server.DefaultReadBufferSize,
				Handler:        serverHandler,
				HBConfig:       hbConfig}

			s := impl.NewServer(&serverConfig)

			s.ListenAndServe(publicAddress, internalAddress)
		}}

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&env, "env", "", "environment")
	rootCmd.Execute()

}

func initConfig() {
	_, err := config.Load(env)
	if err != nil {
		panic(fmt.Sprintf("failed to load config file: %s", env))
	}
}
