package cmd

import (
	"context"
	"fmt"
	"qpush/modules/config"
	"qpush/modules/logger"
	"qpush/modules/rabbitmq"
	"qpush/server"
	"sync"
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
	queue         = "push-queue"
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
					fmt.Println("quit for channel close")
					return
				}

				msgCh <- &msg

			case <-ctx.Done():
				return
			}

		}
	}}

func getMsgs() <-chan amqp.Delivery {

	return rabbitmq.GetMsgs(conf.RabbitMQ, queue, prefetchCount)

}

func handleMsg() {

	defer cancelFunc()

	conns := make(map[string]*qrpc.Connection)
	for _, serverAddr := range conf.Servers {
		conn, err := qrpc.NewConnection(serverAddr, qrpc.ConnectionConfig{}, nil)
		if err != nil {
			panic(err)
		}
		conns[serverAddr] = conn
	}

	for {
		select {
		case d := <-msgCh:
			logger.Debug(string(d.Body))

			var wg sync.WaitGroup
			for _, serverAddr := range conf.Servers {
				conn := conns[serverAddr]
				_, resp, err := conn.Request(server.PushCmd, qrpc.NBFlag, d.Body)
				if err != nil {
					logger.Error("Request", err)
					return
				}
				qrpc.GoFunc(&wg, func() {
					frame := resp.GetFrame()
					if frame == nil {
						logger.Error("GetFrame nil")
						cancelFunc()
					}
				})
			}
			wg.Wait()
			d.Ack(false)
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 5):
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
