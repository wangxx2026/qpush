package agentcmd

import (
	"encoding/json"
	"fmt"
	"qpush/client"
	cimpl "qpush/client/impl"
	"qpush/modules/config"
	"qpush/modules/logger"
	"qpush/server"
	"runtime"
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

var scheduledPushCmd = &cobra.Command{
	Use:   "scheduled_push",
	Short: "get messages to send and push to server",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()

		msgs := prePareQueue()

		for {
			select {
			case msg := <-msgs:
				go handleMsg(&msg)
			case <-time.After(time.Second * 5):
				if len(msgs) == 0 {
					for {
						msgs = prePareQueue()
						if msgs == nil {
							logger.Error("nil from prePareQueue")
							time.Sleep(time.Second)
						} else {
							break
						}
					}

				}
			}

		}
	}}

func prePareQueue() <-chan amqp.Delivery {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		logger.Error("failed to rabbitmq", err)
		return nil
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.Error("failed to open rabbitmq channel", err)
		return nil
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(
		"push_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		logger.Error("failed to QueueDeclare", err)
		return nil
	}
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		logger.Error("failed to set QoS", err)
		return nil
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	return msgs
}

func handleMsg(d *amqp.Delivery) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("recovered from panic in handleMsg", err)
		}
		runtime.GC() //trigger connection close
	}()

	cmd := client.PushCmd{}
	err := json.Unmarshal(d.Body, &cmd)
	if err != nil {
		logger.Error("invalid message", err)
		return
	}

	var routinesGroup sync.WaitGroup
	agent := cimpl.NewAgent()
	for _, serverAddr := range conf.Servers {
		s := serverAddr
		go func() {
			conn := agent.Dial(s)
			if conn == nil {
				logger.Error("failed to dial")
				panic("failed to dial")
			}

			pushCmd := &client.PushCmd{MsgID: 1, Title: "", Content: ""}
			_, err := conn.SendCmdBlocking(server.PushCmd, pushCmd)
			if err != nil {
				logger.Error("SendCmdBlocking failed:", err)
				panic("SendCmdBlocking failed")
			}
			routinesGroup.Done()
		}()
	}
	routinesGroup.Wait()

	d.Ack(false)
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
	rootCmd.AddCommand(scheduledPushCmd)
	scheduledPushCmd.Flags().StringVar(&env, "env", "", "environment")
}
