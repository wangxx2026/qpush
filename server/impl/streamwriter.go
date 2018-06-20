package impl

import (
	"encoding/binary"
	"net"
)

// StreamWriter write data to socket
type StreamWriter struct {
	conn net.Conn
}

// NewStreamWriter new instance
func NewStreamWriter(conn net.Conn) *StreamWriter {
	return &StreamWriter{conn: conn}
}

// WriteUint32 writes uint32
func (w *StreamWriter) WriteUint32(n uint32) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, n)
	return w.WriteBytes(buf)
}

// WriteBytes writes bytes
func (w *StreamWriter) WriteBytes(bytes []byte) error {
	size := len(bytes)

	offset := 0
	for {
		nw, err := w.conn.Write(bytes[offset:])
		if err != nil {
			return err
		}
		offset += nw
		if offset >= size {
			return nil
		}
	}
}
