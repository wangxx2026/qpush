package stream

import (
	"net"
	"time"
)

// Writer write data to socket
type Writer struct {
	conn    net.Conn
	timeout int
}

const (
	// WriteNoTimeout will never timeout
	WriteNoTimeout = -1
)

// NewWriter new instance
func NewWriter(conn net.Conn) *Writer {
	return &Writer{conn: conn, timeout: WriteNoTimeout}
}

// NewWriterWithTimeout new instance with timeout
func NewWriterWithTimeout(conn net.Conn, timeout int) *Writer {
	return &Writer{conn: conn, timeout: timeout}
}

// WriteBytes writes bytes
func (w *Writer) Write(bytes []byte) error {
	size := len(bytes)

	offset := 0

	if w.timeout > 0 {
		w.conn.SetWriteDeadline(time.Now().Add(time.Duration(w.timeout) * time.Second))
	}
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
