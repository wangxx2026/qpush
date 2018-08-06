package cmd

import (
	"context"
	"fmt"
	"qpush/modules/config"
	"qpush/modules/logger"
	"qpush/modules/rabbitmq"
	"qpush/server"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
	"github.com/zhiqiangxu/qrpc"
)

// this file implements scheduled push service

var (
	env             string
	conf            *config.Value
	msgCh           = make(chan *amqp.Delivery)
	ctx, cancelFunc = context.WithCancel(context.Background())
)

const (
	prefetchCount = 100
	pushTimeout   = 30
)

var queuePushCmd = &cobra.Command{
	Use:   "queue_push",
	Short: "get messages to send and push to server",
	Run: func(cmd *cobra.Command, args []string) {

		defer cancelFunc()

		initConfig()

		msgs := getMsgs()
		go handleMsg()

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
	}}

func getMsgs() <-chan amqp.Delivery {

	return rabbitmq.GetMsgs(conf.RabbitMQ, conf.PushQueue, prefetchCount)

}

func initconnect(conns map[string]*qrpc.Connection) {
	for _, serverAddr := range conf.Servers {
		conn, err := qrpc.NewConnection(serverAddr, qrpc.ConnectionConfig{}, nil)
		if err != nil {
			panic(err)
		}
		conns[serverAddr] = conn
	}
}

func handleMsg() {

	defer cancelFunc()

	conns := make(map[string]*qrpc.Connection)
	initconnect(conns)

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
					frame := resp.GetFrame()
					logger.Debug("after GetFrame")
					if frame == nil {
						logger.Error(uuid, addr, "GetFrame nil", addr)
						atomic.StoreInt32(&failed, 1)
						return
					}
					logger.Info(uuid, addr, "push resp", string(frame.Payload))
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
