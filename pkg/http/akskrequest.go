package http

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"qpush/pkg/config"
	"strings"
	"time"
)

const (
	// PostMethod is the only method currently supported by gw
	PostMethod = "POST"
	// DefaultTimeout is default timeout for aksk request
	DefaultTimeout = 3 * time.Second
)

func getAuthorization(method string, uri string, contentType string, body []byte) string {

	conf := config.Get()

	// 第一行Method Path
	firstLine := method + " " + uri + "\n"
	secondLine := "Host: " + conf.GwHost + "\n"
	thirdLine := "Content-Type: " + contentType + "\n\n"

	str := firstLine + secondLine + thirdLine + string(body)

	//hmac ,use sha1
	key := []byte(conf.Sk)
	h := hmac.New(sha1.New, key)

	io.WriteString(h, str)

	encodeString := base64.URLEncoding.EncodeToString(h.Sum(nil))

	code := "mt " + conf.Ak + ":" + encodeString
	return code
}

// DoAkSkRequest send an aksk request and returns response
func DoAkSkRequest(method string, uri string, data interface{}) ([]byte, error) {

	conf := config.Get()

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: DefaultTimeout}
	req, err := http.NewRequest(method, "http://"+conf.GwHost+uri, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	contentType := "application/json"
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", getAuthorization(method, uri, contentType, body))

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	response, err := ioutil.ReadAll(resp.Body)

	return response, err

}
