package http

import (
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
