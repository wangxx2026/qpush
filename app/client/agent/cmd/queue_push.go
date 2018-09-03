package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"qpush/pkg/config"
	"qpush/pkg/logger"
	"qpush/pkg/rabbitmq"
	"qpush/pkg/tail"
	"qpush/server"
	"qpush/server/internalcmd"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"net/http"
	// enable pprof for queue_push
	_ "net/http/pprof"

	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
	"github.com/zhiqiangxu/qrpc"
)

// this file implements scheduled push service

var (
	env  string
	conf *config.Value
)

const (
	prefetchCount  = 100
	writeRespBatch = 100
)

var queuePushCmd = &cobra.Command{
	Use:   "queue_push",
	Short: "get messages to send and push to server",
	Run: func(cmd *cobra.Command, args []string) {

		initConfig()

		srv := &http.Server{Addr: "0.0.0.0:8081"}
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/gc":
				runtime.GC()
				io.WriteString(w, "gc ok\n")
				return
			}
		})
		if conf.AgentLog != "" {
			tail.Attach2Http(http.DefaultServeMux, "/logs", "/wslogs", conf.AgentLog)
		}

		go srv.ListenAndServe()
		defer srv.Close()

		for i := 0; i < 10; i++ {
			logger.Info("Round", i)
			msgCh := make(chan *amqp.Delivery)
			logCh := make(chan []byte)
			ctx, cancelFunc := context.WithCancel(context.Background())
			var wg sync.WaitGroup

			qrpc.GoFunc(&wg, func() {
				handleMQ(ctx, cancelFunc, msgCh)
			})
			qrpc.GoFunc(&wg, func() {
				handleMsg(ctx, cancelFunc, msgCh, logCh)
			})
			qrpc.GoFunc(&wg, func() {
				handleWriteMQ(ctx, cancelFunc, logCh)
			})

			wg.Wait()
		}

	}}

func initconnect(conns map[string]*qrpc.Connection) {
	for _, serverAddr := range conf.Servers {
		conn, err := qrpc.NewConnection(serverAddr, qrpc.ConnectionConfig{}, nil)
		if err != nil {
			panic(err)
		}
		conns[serverAddr] = conn
	}
}

func closeconns(conns map[string]*qrpc.Connection) {
	for _, conn := range conns {
		conn.Close()
	}
}

func handleMQ(ctx context.Context, cancelFunc context.CancelFunc, msgCh chan *amqp.Delivery) {
	defer cancelFunc()

	msgs, mqconn := getMsgs()
	if mqconn == nil {
		logger.Error("getMsgs fail")
		return
	}
	defer mqconn.Close()

	for {
		select {
		case msg, ok := <-msgs:

			if !ok {
				logger.Error("quit for channel close")
				return
			}

			msgCh <- &msg

		case <-ctx.Done():
			return
		}

	}
}

func handleMsg(ctx context.Context, cancelFunc context.CancelFunc, msgCh <-chan *amqp.Delivery, logCh chan<- []byte) {

	defer cancelFunc()

	conns := make(map[string]*qrpc.Connection)
	initconnect(conns)
	defer closeconns(conns)

	for {
		select {
		case d := <-msgCh:
			select {
			case <-ctx.Done():
				return
			default:
			}
			uuid := qrpc.PoorManUUID()
			logger.Info(uuid, string(d.Body))

			msgwg := sync.WaitGroup{}
			failed := int32(0)
			for _, serverAddr := range conf.Servers {
				addr := serverAddr
				conn := conns[addr]
				_, resp, err := conn.Request(server.PushCmd, qrpc.NBFlag, d.Body)
				if err != nil {
					logger.Error(uuid, addr, "Request err", err)
					return
				}

				qrpc.GoFunc(&msgwg, func() {
					logger.Debug("before GetFrame")
					frame, err := resp.GetFrame()
					logger.Debug("after GetFrame")
					if err != nil {
						logger.Error(uuid, addr, "GetFrame fail", err)
						atomic.StoreInt32(&failed, 1)
						return
					}
					logger.Info(uuid, addr, "push resp", string(frame.Payload))
					select {
					case logCh <- frame.Payload:
					case <-ctx.Done():
					}

				})
			}
			go func() {
				logger.Debug("before msgwg wait")
				msgwg.Wait()
				logger.Debug("after msgwg wait")
				if atomic.LoadInt32(&failed) != 0 {
					logger.Error(uuid, "resp fail")
					cancelFunc()
					return
				}

				err := d.Ack(false)
				if err == nil {
					logger.Info(uuid, "done")
				} else {
					logger.Info(uuid, "ack error", err, "quit")
					cancelFunc()
				}
			}()

		case <-ctx.Done():
			return
		case <-time.After(time.Second * 50000):
			fmt.Println("quit for idle")
			return
		}
	}
}

func handleWriteMQ(ctx context.Context, cancelFunc context.CancelFunc, logCh <-chan []byte) {

	defer cancelFunc()

	batchCh := make(chan []internalcmd.PushResp)
	go func() {
		defer cancelFunc()
		for {
			select {
			case <-ctx.Done():
				return
			case respes := <-batchCh:
				bytes, err := json.Marshal(respes)
				respes = nil
				if err != nil {
					logger.Error("marshal respes fail", err)
					continue
				}
				err = rabbitmq.ProduceMsgKeepAlive(conf.RabbitMQ, "", "push-resp", string(bytes))
				if err != nil {
					logger.Error("ProduceMsgKeepAlive fail", err)
				}
				logger.Info("push-resp OK", string(bytes))
			}
		}
	}()
	var respes []internalcmd.PushResp

	for {
		select {
		case logBytes := <-logCh:
			var pushResp internalcmd.PushResp
			err := json.Unmarshal(logBytes, &pushResp)
			if err != nil {
				logger.Error("PushResp unmarshal fail", err)
				continue //ignore error
			}
			if pushResp.OK == 0 && pushResp.NG == 0 {
				logger.Debug("ignore all zero", string(logBytes))
				continue
			}
			respes = append(respes, pushResp)
			if len(respes) > writeRespBatch {
				// TODO fix dup
				select {
				case batchCh <- respes:
				case <-ctx.Done():
					return
				}
				respes = nil
			}

		case <-time.After(time.Second * 60):
			if len(respes) > 0 {
				// TODO fix dup
				select {
				case batchCh <- respes:
				case <-ctx.Done():
					return
				}
				respes = nil
			}
		case <-ctx.Done():
			return
		}
	}
}

func getMsgs() (<-chan amqp.Delivery, *amqp.Connection) {
	return rabbitmq.GetMsgs(conf.RabbitMQ, conf.PushQueue, prefetchCount)
}

func initConfig() {
	var err error
	conf, err = config.Load(env)
	if err != nil {
		panic(fmt.Sprintf("failed to load config file: %s", env))
	}
}

func init() {
	rootCmd.AddCommand(queuePushCmd)
	queuePushCmd.Flags().StringVar(&env, "env", "", "environment")
}
