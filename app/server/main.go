package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"qpush/client"
	"qpush/pkg/config"
	"qpush/pkg/logger"
	"qpush/server"
	"qpush/server/cmd"
	"qpush/server/internalcmd"
	"runtime"
	"strconv"
	"syscall"
	"time"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/zhiqiangxu/qrpc"
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

			gaugeMetric := kitprometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Namespace: "qpush",
				Subsystem: "server",
				Name:      "gauge_result",
				Help:      "The gauge result per app.",
			}, []string{"appid", "kind"})
			summaryMetric := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
				Namespace: "qpush",
				Subsystem: "server",
				Name:      "summary_result",
				Help:      "request latency.",
			}, []string{"method", "error"})

			handler := qrpc.NewServeMux()
			handler.Handle(server.LoginCmd, cmd.NewLoginCmd(gaugeMetric))
			handler.Handle(server.HeartBeatCmd, &cmd.HeartBeatCmd{})
			handler.Handle(server.AckCmd, cmd.NewAckCmd())

			internalHandler := qrpc.NewServeMux()
			internalHandler.Handle(server.PushCmd, &internalcmd.PushCmd{})
			internalHandler.Handle(server.ListGUIDCmd, &internalcmd.ListGUIDCmd{})
			internalHandler.Handle(server.ExecCmd, &internalcmd.ExecCmd{})
			internalHandler.Handle(server.KillCmd, &internalcmd.KillCmd{})
			internalHandler.Handle(server.RelayCmd, &internalcmd.RelayCmd{})
			internalHandler.Handle(server.CheckGUIDCmd, &internalcmd.CheckGUIDCmd{})

			bindings := []qrpc.ServerBinding{
				qrpc.ServerBinding{Addr: publicAddr, Handler: handler, DefaultReadTimeout: 10 /*second*/, LatencyMetric: summaryMetric},
				qrpc.ServerBinding{Addr: internalAddr, Handler: internalHandler, LatencyMetric: summaryMetric}}

			qserver := qrpc.NewServer(bindings)

			go func() {
				srv := &http.Server{Addr: "0.0.0.0:8080"}
				http.Handle("/metrics", promhttp.Handler())
				http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

					switch r.URL.Path {
					case "/gc":
						runtime.GC()
						io.WriteString(w, "gc ok\n")
						return
					}

					if config.Get().Env == config.ProdEnv {
						return
					}

					logger.Debug("tsetxxx", r.URL.Path)
					switch r.URL.Path {
					case "/":
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
					case "/kill":
						appid := r.URL.Query().Get("appid")
						id, err := strconv.Atoi(appid)
						if err != nil {
							panic(err)
						}
						guid := r.URL.Query().Get("guid")
						qserver.WalkConnByID(0, []string{server.GetAppGUID(id, guid)}, func(w qrpc.FrameWriter, ci *qrpc.ConnectionInfo) {
							ci.SC.Close()
						})
					}

					runtime.GC()
					io.WriteString(w, "ok\n")
				})
				fmt.Println(srv.ListenAndServe())

			}()
			go func() {
				err := qserver.ListenAndServe()
				if err != nil {
					fmt.Println("ListenAndServe", err)
				}
			}()

			quitChan := make(chan os.Signal, 1)
			signal.Notify(quitChan, os.Interrupt, os.Kill, syscall.SIGTERM)

			<-quitChan
			err := qserver.Shutdown()
			logger.Info("Shutdown")
			if err != nil {
				logger.Error("Shutdown", err)
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
