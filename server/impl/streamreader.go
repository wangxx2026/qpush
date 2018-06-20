package impl

import (
	"encoding/binary"
	"io"
	"net"
)

// StreamReader read data from socket
type StreamReader struct {
	conn net.Conn
}

// NewStreamReader creates a StreamReader instance
func NewStreamReader(conn net.Conn) *StreamReader {
	return &StreamReader{conn: conn}
}

// ReadUint32 read uint32 from socket
func (r *StreamReader) ReadUint32() (uint32, error) {
	bytes := make([]byte, 4)
	_, err := io.ReadFull(r.conn, bytes)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(bytes), nil
}

// ReadBytes read bytes
func (r *StreamReader) ReadBytes(bytes []byte) error {
	_, err := io.ReadFull(r.conn, bytes)
	if err != nil {
		return err
	}

	return nil
}
