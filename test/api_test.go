package test

import (
	"fmt"
	"qpush/pkg/config"
	"qpush/pkg/http"
	"testing"
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
