package agentcmd

import (
	"encoding/json"
	"fmt"
	"qpush/client"
	cimpl "qpush/client/impl"
	"qpush/modules/config"
	"qpush/modules/logger"
	"qpush/modules/rabbitmq"
	"qpush/server"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
)

// this file implements scheduled push service

var (
	env  string
	conf *config.Value
)

const (
	queue         = "push-queue"
	prefetchCount = 1
)

var queuePushCmd = &cobra.Command{
	Use:   "queue_push",
	Short: "get messages to send and push to server",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()

		msgs := getMsgs()

		for {
			select {
			case msg, ok := <-msgs:

				if !ok {
					panic("quit for channel close")
				}

				go handleMsg(&msg)

			case <-time.After(time.Second * 5):
				panic("quit for idle")
			}

		}
	}}

func getMsgs() <-chan amqp.Delivery {

	return rabbitmq.GetMsgs(conf.RabbitMQ, queue, prefetchCount)

}

func handleMsg(d *amqp.Delivery) {

	logger.Debug(string(d.Body))

	defer func() {
		if err := recover(); err != nil {
			logger.Error("recovered from panic in handleMsg", err)
		}
	}()

	cmd := client.PushCmd{}
	err := json.Unmarshal(d.Body, &cmd)
	if err != nil {
		logger.Error("invalid message", err)
		return
	}
	logger.Debug(cmd)

	var routinesGroup sync.WaitGroup
	agent := cimpl.NewAgent()
	for _, serverAddr := range conf.Servers {
		s := serverAddr
		routinesGroup.Add(1)
		go func() {
			conn := agent.Dial(s)
			if conn == nil {
				logger.Error("failed to dial")
				panic("failed to dial")
			}

			_, err := conn.SendCmdBlocking(server.PushCmd, &cmd)
			if err != nil {
				logger.Error("SendCmdBlocking failed:", err)
				panic("SendCmdBlocking failed")
			}
			routinesGroup.Done()
		}()
	}
	routinesGroup.Wait()

	// d.Ack(false)
}

func goFunc(routinesGroup *sync.WaitGroup, f func()) {
	routinesGroup.Add(1)
	go func() {
		defer routinesGroup.Done()
		f()
	}()
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
