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
	"qpush/pkg/tail"
	"qpush/server"
	"qpush/server/cmd"
	"qpush/server/internalcmd"
	"runtime"
	"sort"
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
		Args:  cobra.MaximumNArgs(2),
		Run: func(cobraCmd *cobra.Command, args []string) {
			publicAddr, internalAddr := DefaultPublicAddress, DefaultInternalAddress
			slice := []*string{&publicAddr, &internalAddr}
			for i, arg := range args {
				*slice[i] = arg
			}

			defer func() {
				if err := recover(); err != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					logger.Error("main thread panic", err, buf)
				}
			}()

			// online count
			onlineMetric := kitprometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Namespace: "qpush",
				Subsystem: "server",
				Name:      "online_count",
				Help:      "The gauge result per app.",
			}, []string{"appid", "kind"})
			// ok|ng count
			pushCounterMetric := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
				Namespace: "qpush",
				Subsystem: "server",
				Name:      "push_count",
				Help:      "The counter result per app.",
			}, []string{"appid", "kind"})
			// qrpc request count
			requestCountMetric := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
				Namespace: "qpush",
				Subsystem: "server",
				Name:      "request_count",
				Help:      "The counter result per app.",
			}, []string{"method", "error"})
			// qrpc request latency
			requestLatencyMetric := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
				Namespace: "qpush",
				Subsystem: "server",
				Name:      "request_latency",
				Help:      "request latency.",
			}, []string{"method", "error"})

			handler := qrpc.NewServeMux()
			handler.Handle(server.LoginCmd, cmd.NewLoginCmd(onlineMetric, pushCounterMetric))
			handler.Handle(server.HeartBeatCmd, &cmd.HeartBeatCmd{})
			handler.Handle(server.AckCmd, cmd.NewAckCmd())

			internalHandler := qrpc.NewServeMux()
			internalHandler.Handle(server.PushCmd, internalcmd.NewPushCmd(pushCounterMetric))
			internalHandler.Handle(server.ListGUIDCmd, &internalcmd.ListGUIDCmd{})
			internalHandler.Handle(server.ExecCmd, &internalcmd.ExecCmd{})
			internalHandler.Handle(server.KillCmd, &internalcmd.KillCmd{})
			internalHandler.Handle(server.RelayCmd, &internalcmd.RelayCmd{})
			internalHandler.Handle(server.CheckGUIDCmd, &internalcmd.CheckGUIDCmd{})

			bindings := []qrpc.ServerBinding{
				qrpc.ServerBinding{Addr: publicAddr, Handler: handler, DefaultReadTimeout: 10 /*second*/, LatencyMetric: requestLatencyMetric, CounterMetric: requestCountMetric},
				qrpc.ServerBinding{Addr: internalAddr, Handler: internalHandler, LatencyMetric: requestLatencyMetric, CounterMetric: requestCountMetric}}

			qserver := qrpc.NewServer(bindings)

			go func() {
				srv := &http.Server{Addr: "0.0.0.0:8080"}
				http.Handle("/metrics", promhttp.Handler())
				if config.Get().ServerLog != "" {
					tail.Attach2Http(http.DefaultServeMux, "/logs", "/wslogs", config.Get().ServerLog)
				}
				http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

					switch r.URL.Path {
					case "/gc":
						runtime.GC()
						io.WriteString(w, "gc ok2\n")
						return
					case "/listguid":
						var result server.DeviceInfoSlice
						qserver.WalkConn(0, func(w qrpc.FrameWriter, ci *qrpc.ConnectionInfo) bool {
							deviceInfo, ok := ci.GetAnything().(*server.DeviceInfo)
							if ok {
								result = append(result, deviceInfo)
							}

							return true
						})

						sort.Sort(result)
						bytes, _ := json.Marshal(result)
						io.WriteString(w, string(bytes))
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
