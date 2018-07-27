package cmd

import (
	"context"
	"fmt"
	"qpush/modules/config"
	"qpush/modules/logger"
	"qpush/modules/rabbitmq"
	"qpush/server"
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
			case <-time.After(time.Second * 5):
				fmt.Println("quit for idle")
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

			for _, serverAddr := range conf.Servers {
				conn := conns[serverAddr]
				_, resp, err := conn.Request(server.PushCmd, qrpc.NBFlag, d.Body)
				if err != nil {
					panic(err)
				}
				frame := resp.GetFrame()
				if frame == nil {
					panic("no response")
				}
			}

			d.Ack(false)

		case <-ctx.Done():
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
