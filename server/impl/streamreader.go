package impl

import (
	"encoding/binary"
	"io"
	"net"
	"time"
)

// StreamReader read data from socket
type StreamReader struct {
	conn    net.Conn
	timeout int
}

const (
	// ReadNoTimeout will never timeout
	ReadNoTimeout = -1
)

// NewStreamReader creates a StreamReader instance
func NewStreamReader(conn net.Conn) *StreamReader {
	return &StreamReader{conn: conn, timeout: ReadNoTimeout}
}

// NewStreamReaderWithTimeout allows specify timeout
func NewStreamReaderWithTimeout(conn net.Conn, timeout int) *StreamReader {
	return &StreamReader{conn: conn, timeout: timeout}
}

// ReadUint32 read uint32 from socket
func (r *StreamReader) ReadUint32() (uint32, error) {
	bytes := make([]byte, 4)
	if r.timeout > 0 {
		r.conn.SetReadDeadline(time.Now().Add(time.Duration(r.timeout) * time.Second))
	}
	_, err := io.ReadFull(r.conn, bytes)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(bytes), nil
}

// ReadBytes read bytes
func (r *StreamReader) ReadBytes(bytes []byte) error {
	if r.timeout > 0 {
		r.conn.SetReadDeadline(time.Now().Add(time.Duration(r.timeout) * time.Second))
	}
	_, err := io.ReadFull(r.conn, bytes)
	if err != nil {
		return err
	}

	return nil
}
