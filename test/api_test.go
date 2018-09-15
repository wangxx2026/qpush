package test

import (
	"context"
	"encoding/json"
	"fmt"
	"qpush/client"
	"qpush/pkg/config"
	"qpush/pkg/http"
	"qpush/server"
	"sync"
	"testing"
	"time"

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

func BenchmarkUpdateUserInfo(b *testing.B) {

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
			dataW := make(map[string]interface{})
			data := make(map[string]interface{})
			dev := make(map[string]interface{})
			dev["android_id"] = "23af56760207669d"
			dev["sn"] = "90c14c34"
			dev["os"] = "Xiaomi"
			dev["os_version"] = "os_version1"
			dev["os_device"] = "os_device1"
			data["guid"] = "8bad52ca-d051-5c84-a10b-93f230c94a45"
			data["app_id"] = 1008
			data["app_key"] = "ddddddd"
			data["device_info"] = dev
			data["device_token1"] = "dddddddddddfgggxxxxxxx"
			data["device_token2"] = "8VvH+lqswYwQSrAlY5xU6RcVtlRwYsbQCTr1NnrUggggxxxgg"
			data["channel"] = "ios"
			data["open_notice"] = true
			data["open_id"] = "1231231"
			//data["idempotent"] = "abcdefghijklnnnngggg"
			data["idempotent"] = "abcdefghijklnnnnggggxxxggg"
			data["version"] = "vt1.0.1"
			data["open_id"] = "open_idxxxxxx"
			dataW["info"] = data

			_, err := http.DoAkSkRequest(http.PostMethod, "/v1/pushaksk/updateuserinfo", dataW)

			if err != nil {
				b.Fatalf("DoAkSkRequest err: %v", err)
			}
		}

	})
}

func BenchmarkHeartBeat(b *testing.B) {
	ctx := context.Background()
	endpoints := []string{"localhost:8888"}
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

func BenchmarkPush(b *testing.B) {
	ctx := context.Background()
	endpoints := []string{"localhost:8890"}
	api := qrpc.NewAPI(endpoints, qrpc.ConnectionConfig{}, nil)

	cmd := client.PushCmd{Msg: client.Msg{MsgID: "testxx", Title: "title", Content: "content"}, AppID: 1008}
	bytes, _ := json.Marshal(cmd)

	b.SetParallelism(500)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := api.Call(ctx, server.PushCmd, bytes)
			if err != nil {
				b.Fatalf("heartbeat fail:%v", err)
			}
		}
	})
}

func TestConsumers(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		qrpc.GoFunc(&wg, func() {
			conn, err := qrpc.NewConnection("localhost:8888", qrpc.ConnectionConfig{}, func(conn *qrpc.Connection, frame *qrpc.Frame) {
				fmt.Println(string(frame.Payload))
			})
			if err != nil {
				panic("err")
			}
			loginCmd := client.LoginCmd{GUID: "", AppID: 1, AppKey: ""}
			bytes, _ := json.Marshal(loginCmd)
			conn.Request(server.LoginCmd, 0, bytes)
			time.Sleep(time.Second * 5)
			fmt.Println(conn)
		})
	}
	wg.Wait()
}
