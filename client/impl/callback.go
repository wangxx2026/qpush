package impl

import "push-msg/server"

// OnResponse is data struct for callback
type OnResponse struct {
	f func(uint64, server.Cmd, []byte) error
}

// NewCallBack creates an instance
func NewCallBack(cb func(uint64, server.Cmd, []byte) error) *OnResponse {
	return &OnResponse{f: cb}
}

// Call is called when message arived
func (cb *OnResponse) Call(requestID uint64, cmd server.Cmd, bytes []byte) error {
	return cb.f(requestID, cmd, bytes)
}
