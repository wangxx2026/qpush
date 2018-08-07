package client

import (
	"encoding/json"
	"qpush/pkg/logger"

	"github.com/zhiqiangxu/qrpc"
)

// Connection for json rpc
type Connection struct {
	*qrpc.Connection
}

// NewConnection is a wrapper around qrpc.NewConnection
func NewConnection(addr string, conf qrpc.ConnectionConfig, f func(*Connection, *qrpc.Frame)) (*Connection, error) {
	conn := &Connection{}
	c, err := qrpc.NewConnection(addr, conf, func(c *qrpc.Connection, frame *qrpc.Frame) {
		f(conn, frame)
	})
	if err != nil {
		logger.Error("NewConnection fail", err)
		return nil, err
	}
	conn.Connection = c
	return conn, nil
}

// StreamRequest for json encode
func (conn *Connection) StreamRequest(cmd qrpc.Cmd, flags qrpc.FrameFlag, payload interface{}) (*StreamWriter, qrpc.Response, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	w, resp, err := conn.Connection.StreamRequest(cmd, flags, bytes)
	return &StreamWriter{w}, resp, err
}

// Request for json encode
func (conn *Connection) Request(cmd qrpc.Cmd, flags qrpc.FrameFlag, payload interface{}) (uint64, qrpc.Response, error) {

	bytes, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, err
	}

	return conn.Connection.Request(cmd, flags, bytes)
}

// StreamWriter for json
type StreamWriter struct {
	qrpc.StreamWriter
}

// WriteBytes for json
func (w *StreamWriter) WriteBytes(data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.StreamWriter.WriteBytes(bytes)
	return nil
}
