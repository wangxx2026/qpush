package rabbitmq

import (
	"qpush/modules/logger"

	"github.com/streadway/amqp"
)

// GetMsgs returns message channel
func GetMsgs(url string, topic string, prefetchCount int) <-chan amqp.Delivery {
	conn, err := amqp.Dial(url)
	if err != nil {
		logger.Error("failed to dial rabbitmq", url, err)
		return nil
	}

	ch, err := conn.Channel()
	if err != nil {
		logger.Error("failed to open rabbitmq channel", err)
		return nil
	}
	q, err := ch.QueueDeclare(
		topic, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		logger.Error("failed to QueueDeclare", err)
		return nil
	}
	err = ch.Qos(
		prefetchCount, // prefetch count
		0,             // prefetch size
		false,         // global
	)
	if err != nil {
		logger.Error("failed to set QoS", err)
		return nil
	}
	msgs, err := ch.Consume(
		q.Name,  // queue
		"qpush", // consumer
		false,   // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	return msgs
}
