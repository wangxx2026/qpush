package http

import (
	"encoding/json"
	"fmt"
	"os"
	"qpush/pkg/config"
	"testing"
)

func TestAksk(t *testing.T) {

	os.Chdir("/Users/xuzhiqiang/.gvm/pkgsets/go1.10/global/src/qpush")
	_, err := config.Load("dev")
	if err != nil {
		t.Fatal(err)
	}

	data := map[string]interface{}{"app_id": 1008, "app_key": "ddddddd", "guid": "guid"}

	resp, err := DoAkSkRequest("POST", "/v1/pushaksk/offlinemsg", data)

	fmt.Println(string(resp), err)
}

func TestAuthorization(t *testing.T) {
	os.Chdir("/Users/xuzhiqiang/.gvm/pkgsets/go1.10/global/src/qpush")
	config.Load("prod")

	data := map[string]interface{}{"app_id": 1001, "app_key": "UwMTA1Nw", "guid": "c04792e2-97cf-4371-b547-cccd4fdac210"}
	bytes, _ := json.Marshal(data)
	fmt.Println(getAuthorization("POST", "/v1/pushaksk/offlinemsg", "application/json", bytes))
}
