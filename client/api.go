package client

import (
	"context"
	"encoding/json"

	"github.com/zhiqiangxu/qrpc"
)

// API for non blocking roundtrip calls
type API interface {
	Call(ctx context.Context, cmd qrpc.Cmd, payload interface{}, result interface{}) error
	CallOne(ctx context.Context, endpoint string, cmd qrpc.Cmd, payload interface{}, result interface{}) error
	CallAll(ctx context.Context, cmd qrpc.Cmd, payload interface{}) (map[string]*qrpc.APIResult, error)
}

type defaultAPI struct {
	api qrpc.API
}

// NewAPI creates an API instance
func NewAPI(endpoints []string, conf qrpc.ConnectionConfig, weights []int) API {

	return &defaultAPI{api: qrpc.NewAPI(endpoints, conf, weights)}
}

func (a *defaultAPI) Call(ctx context.Context, cmd qrpc.Cmd, payload interface{}, result interface{}) error {

	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	frame, err := a.api.Call(ctx, cmd, bytes)
	if err != nil {
		return err
	}
	return json.Unmarshal(frame.Payload, result)
}

func (a *defaultAPI) CallOne(ctx context.Context, endpoint string, cmd qrpc.Cmd, payload interface{}, result interface{}) error {

	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	frame, err := a.api.CallOne(ctx, endpoint, cmd, bytes)
	if err != nil {
		return nil
	}
	return json.Unmarshal(frame.Payload, result)
}

func (a *defaultAPI) CallAll(ctx context.Context, cmd qrpc.Cmd, payload interface{}) (map[string]*qrpc.APIResult, error) {

	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return a.api.CallAll(ctx, cmd, bytes), nil
}
