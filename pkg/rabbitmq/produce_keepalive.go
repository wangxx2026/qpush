package rabbitmq

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/streadway/amqp"
)

type keepAlivedPublisher struct {
	m     sync.Mutex
	conns map[string]*keepAlivedConn
}

type keepAlivedConn struct {
	m      sync.Mutex
	conn   *amqp.Connection
	chs    map[string]*amqp.Channel
	closed int32
}

func newKeepAlivedConn(url string) (*keepAlivedConn, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	return &keepAlivedConn{conn: conn, chs: make(map[string]*amqp.Channel)}, nil
}

func (kconn *keepAlivedConn) Publish(exchange, routingKey, msg string) (err error) {
	kconn.m.Lock()
	defer kconn.m.Unlock()

	key := fmt.Sprintf("%s:%s", exchange, routingKey)
	ch, ok := kconn.chs[key]
	if !ok {
		ch, err = kconn.conn.Channel()
		if err != nil {
			return
		}

		if routingKey != "" {
			_, err = ch.QueueDeclare(
				routingKey, // name
				true,       // durable
				false,      // delete when unused
				false,      // exclusive
				false,      // no-wait
				nil,        // arguments
			)
			if err != nil {
				return
			}
		}

		kconn.chs[key] = ch
	}

	return ch.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(msg),
		})

}

func (kconn *keepAlivedConn) Close() {
	kconn.m.Lock()
	defer kconn.m.Unlock()

	if atomic.LoadInt32(&kconn.closed) != 0 {
		return
	}
	atomic.StoreInt32(&kconn.closed, 1)

	for _, ch := range kconn.chs {
		ch.Close()
	}

	kconn.conn.Close()
}

func (kconn *keepAlivedConn) IsClosed() bool {
	return atomic.LoadInt32(&kconn.closed) != 0
}

var (
	publisher = keepAlivedPublisher{conns: make(map[string]*keepAlivedConn)}
)

// ProduceMsgKeepAlive will keep the underlying connection alive
func ProduceMsgKeepAlive(url, exchange, routingKey, msg string) (err error) {

	for i := 1; i <= 3; i++ {
		conn, err := getConn(url)
		if err != nil {
			continue
		}

		err = conn.Publish(exchange, routingKey, msg)
		if err != nil {
			conn.Close()
			continue
		}
		return nil
	}

	return
}

func getConn(url string) (*keepAlivedConn, error) {
	publisher.m.Lock()
	defer publisher.m.Unlock()

	conn, ok := publisher.conns[url]
	if ok {
		if !conn.IsClosed() {
			return conn, nil
		}
	}

	conn, err := newKeepAlivedConn(url)
	if err != nil {
		return nil, err
	}
	publisher.conns[url] = conn
	return conn, nil

}
