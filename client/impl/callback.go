package impl

// OnResponse is data struct for callback
type OnResponse struct {
	cb func(requestID uint64, bytes []byte) error
}

// NewCallBack creates an instance
func NewCallBack(cb func(requestID uint64, bytes []byte) error) *OnResponse {
	return &OnResponse{cb: cb}
}

// Call is called when message arived
func (cb *OnResponse) Call(requestID uint64, bytes []byte) error {
	return cb.Call(requestID, bytes)
}
