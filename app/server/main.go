package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"qpush/client"
	"qpush/modules/config"
	"qpush/server"
	"qpush/server/cmd"
	"qpush/server/internalcmd"
	"runtime"
	"time"

	"github.com/zhiqiangxu/qrpc"

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
				publicAddr   string
				internalAddr string
			)
			if len(args) > 0 {
				publicAddr = args[0]
			} else {
				publicAddr = DefaultPublicAddress
			}

			if len(args) > 1 {
				internalAddr = args[1]
			} else {
				internalAddr = DefaultInternalAddress
			}

			handler := qrpc.NewServeMux()
			handler.Handle(server.LoginCmd, &cmd.LoginCmd{})
			handler.Handle(server.HeartBeatCmd, &cmd.HeartBeatCmd{})
			handler.Handle(server.AckCmd, cmd.NewAckCmd())

			internalHandler := qrpc.NewServeMux()
			internalHandler.Handle(server.PushCmd, &internalcmd.PushCmd{})
			internalHandler.Handle(server.ListGUIDCmd, &internalcmd.ListGUIDCmd{})
			internalHandler.Handle(server.ExecCmd, &internalcmd.ExecCmd{})
			internalHandler.Handle(server.KillCmd, &internalcmd.KillCmd{})

			bindings := []qrpc.ServerBinding{
				qrpc.ServerBinding{Addr: publicAddr, Handler: handler},
				qrpc.ServerBinding{Addr: internalAddr, Handler: internalHandler}}

			qserver := qrpc.NewServer(bindings)

			go func() {
				srv := &http.Server{Addr: "0.0.0.0:8080"}
				http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

					if config.Get().Env == config.ProdEnv {
						return
					}

					if r.URL.Path != "/" {
						return
					}

					id := "1"
					title := "test title"
					content := "test content"

					msg := client.Msg{
						MsgID: id, Title: title, Content: content}
					payload, _ := json.Marshal(msg)
					pushID := qserver.GetPushID()

					qserver.WalkConn(0, func(w qrpc.FrameWriter, ci *qrpc.ConnectionInfo) bool {
						w.StartWrite(pushID, server.ForwardCmd, qrpc.PushFlag)
						w.WriteBytes(payload)
						w.EndWrite()
						return true
					})

					runtime.GC()
					io.WriteString(w, "ok\n")
				})
				fmt.Println(srv.ListenAndServe())
			}()
			err := qserver.ListenAndServe()
			if err != nil {
				fmt.Println("ListenAndServe", err)
			}
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
