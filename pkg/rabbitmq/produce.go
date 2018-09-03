package rabbitmq

import (
	"github.com/streadway/amqp"
)

// ProduceMsg produce msg to topic
func ProduceMsg(url, topic, msg string) error {
	conn, err := amqp.Dial(url)

	if err != nil {
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()

	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		topic, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	return ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(msg),
		})
}
