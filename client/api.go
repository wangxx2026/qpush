package client

import (
	"context"
	"encoding/json"

	"github.com/zhiqiangxu/qrpc"
)

// API for non blocking roundtrip calls
type API interface {

	// call random endpoint without decoding
	CallForFrame(ctx context.Context, cmd qrpc.Cmd, payload interface{}) (*qrpc.Frame, error)
	// call random endpoint
	Call(ctx context.Context, cmd qrpc.Cmd, payload interface{}, result interface{}) error
	// call specified endpoint without decoding
	CallOneForFrame(ctx context.Context, endpoint string, cmd qrpc.Cmd, payload interface{}) (*qrpc.Frame, error)
	// call specified endpoint
	CallOne(ctx context.Context, endpoint string, cmd qrpc.Cmd, payload interface{}, result interface{}) error
	// call all endpoint
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

	frame, err := a.CallForFrame(ctx, cmd, payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(frame.Payload, result)
}

func (a *defaultAPI) CallForFrame(ctx context.Context, cmd qrpc.Cmd, payload interface{}) (*qrpc.Frame, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return a.api.Call(ctx, cmd, bytes)
}

func (a *defaultAPI) CallOne(ctx context.Context, endpoint string, cmd qrpc.Cmd, payload interface{}, result interface{}) error {

	frame, err := a.CallOneForFrame(ctx, endpoint, cmd, payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(frame.Payload, result)
}

func (a *defaultAPI) CallOneForFrame(ctx context.Context, endpoint string, cmd qrpc.Cmd, payload interface{}) (*qrpc.Frame, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return a.api.CallOne(ctx, endpoint, cmd, bytes)
}

func (a *defaultAPI) CallAll(ctx context.Context, cmd qrpc.Cmd, payload interface{}) (map[string]*qrpc.APIResult, error) {

	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return a.api.CallAll(ctx, cmd, bytes), nil
}
