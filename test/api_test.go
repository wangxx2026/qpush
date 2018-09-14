package test

import (
	"context"
	"fmt"
	"qpush/pkg/config"
	"qpush/pkg/http"
	"qpush/server"
	"testing"

	"github.com/zhiqiangxu/qrpc"
)

const (
	apiOffline = "/v1/pushaksk/offlinemsg"
	apiAck     = "/v1/pushaksk/notifymsg"
)

// Benchmark apiOffline
func BenchmarkApiOffline(b *testing.B) {

	_, err := config.LoadFile("../pkg/config/prod.toml")
	if err != nil {
		panic(fmt.Sprintf("failed to load config file: %v", err))
	}

	b.SetParallelism(500)
	b.RunParallel(func(pb *testing.PB) {

		for pb.Next() {

			// test
			// data := map[string]interface{}{"app_id": 1008, "app_key": "ddddddd", "guid": "guid"}
			// prod
			data := map[string]interface{}{"app_id": 1001, "app_key": "UwMTA1Nw", "guid": "c04792e2-97cf-4371-b547-cccd4fdac210"}

			_, err := http.DoAkSkRequest(http.PostMethod, "/v1/pushaksk/offlinemsg", data)

			if err != nil {
				b.Fatalf("DoAkSkRequest err: %v", err)
			}
		}

	})
}

func BenchmarkHeartBeat(b *testing.B) {
	ctx := context.Background()
	endpoints := []string{"106.14.164.33:8888"}
	api := qrpc.NewAPI(endpoints, qrpc.ConnectionConfig{}, nil)

	b.SetParallelism(500)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := api.Call(ctx, server.HeartBeatCmd, nil)
			if err != nil {
				b.Fatalf("heartbeat fail:%v", err)
			}
		}
	})

}
