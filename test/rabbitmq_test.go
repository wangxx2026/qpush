package test

import (
	"fmt"
	"qpush/pkg/config"
	"qpush/pkg/rabbitmq"
	"testing"
)

func BenchmarkRabbitMQ(b *testing.B) {
	conf, err := config.LoadFile("../pkg/config/dev.toml")
	if err != nil {
		panic(fmt.Sprintf("failed to load config file: %v", err))
	}

	b.SetParallelism(5000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err = rabbitmq.ProduceMsgKeepAlive(conf.RabbitMQ, "", "push-benchmark", "hello world")
			if err != nil {
				b.Fatalf("ProduceMsg err: %v", err)
			}
		}
	})

}
