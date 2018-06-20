package impl

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"push-msg/modules/logger"
	"push-msg/server"
	"sync"
	"syscall"
	"time"
)

// Server is data struct for server
type Server struct {
	connWriteChans sync.Map
	guidConn       sync.Map
	connGUID       sync.Map
	readBufferSize int
	handler        server.Handler
}

const (
	// DefaultAcceptTimeout is the default accept timeout duration
	DefaultAcceptTimeout = 5 * time.Second
)

// NewServer creates a server instance
func NewServer(c *server.Config) *Server {
	var (
		readBufferSize int
		handler        server.Handler
	)
	if c == nil {
		readBufferSize = server.DefaultReadBufferSize
		handler = &ServerHandler{}
	} else {
		readBufferSize = c.ReadBufferSize
		handler = c.Handler
	}

	return &Server{
		readBufferSize: readBufferSize,
		handler:        handler}
}

// ListenAndServe start listen and serve
func (s *Server) ListenAndServe(address string, internalAddress string) error {

	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	done := make(chan bool)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go s.listenAndServe(address, false, done, &wg)
	wg.Add(1)
	go s.listenAndServe(internalAddress, true, done, &wg)

	go s.handleSignal(quitChan, done)
	wg.Wait()

	return nil
}

// Walk walks each connection
func (s *Server) Walk(f func(net.Conn, chan []byte) bool) {
	s.connWriteChans.Range(func(k, v interface{}) bool {
		return f(k.(net.Conn), v.(chan []byte))
	})
}

func (s *Server) handleSignal(quitChan chan os.Signal, done chan bool) {
	<-quitChan
	logger.Info("signal captured")
	close(done)
}

func (s *Server) listenAndServe(address string, internal bool, done chan bool, wg *sync.WaitGroup) {

	defer wg.Done()

	listener, err := net.Listen("tcp", address)

	if err != nil {
		panic(fmt.Sprintf("listen failed: %s", address))
	}

	defer listener.Close()

	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		panic("shouldn't happen")
	}
	for {
		select {
		case <-done:
			return
		default:
		}
		tcpListener.SetDeadline(time.Now().Add(DefaultAcceptTimeout))
		conn, err := listener.Accept()
		if err != nil {
			if opError, ok := err.(*net.OpError); ok && opError.Timeout() {
				// don't log the scheduled timeout
				continue
			}
			logger.Error(fmt.Sprintf("accept failed:%s", err))
			continue
		}

		wg.Add(1)
		go s.handleConnection(conn, internal, done, wg)
	}
}

func (s *Server) handleConnection(conn net.Conn, internal bool, done chan bool, wg *sync.WaitGroup) {

	defer s.closeConnection(conn)
	defer wg.Done()

	writeChan := make(chan []byte, 30)
	s.connWriteChans.Store(conn, writeChan)

	r := NewStreamReader(conn)

	go s.handleWrite(conn, writeChan)

	for {
		select {
		case <-done:
			close(writeChan)
			return
		default:
		}

		size, err := r.ReadUint32()
		if err != nil {
			logger.Error(fmt.Sprintf("ReadUint32 failed:%s", err))
			return
		}

		if size < 8 {
			logger.Error("invalid packet")
			return
		}

		payload := make([]byte, size)
		err = r.ReadBytes(payload)
		if err != nil {
			logger.Error(fmt.Sprintf("ReadBytes failed:%s", err))
			return
		}

		requestID := binary.BigEndian.Uint64(payload)

		m := make(map[string]interface{})
		json.Unmarshal(payload[8:], &m)

		var (
			cmd interface{}
			ok  bool
		)
		if cmd, ok = m["cmd"]; !ok {
			logger.Error(fmt.Sprintf("invalid payload:%s", string(payload)))
			return
		}

		params := server.CmdParam{Param: m, Server: s, Conn: conn}

		response, err := s.handler.Call(cmd.(string), internal, &params)
		if err != nil {
			logger.Error("handler.Call return error:%s", err)
			return
		}

		if response == nil {
			continue
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			logger.Error("json.Marshal fail:%s", err)
			return
		}

		logger.Debug("response is ", string(jsonResponse))

		packet := makeResponsePacket(requestID, jsonResponse)
		writeChan <- packet

	}

}

func makeResponsePacket(requestID uint64, json []byte) []byte {
	length := 8 + uint32(len(json))
	buf := make([]byte, length)
	binary.BigEndian.PutUint32(buf, length)
	binary.BigEndian.PutUint64(buf, requestID)
	copy(buf[12:], json)
	return buf
}

func (s *Server) handleWrite(conn net.Conn, writeChann chan []byte) {
	w := NewStreamWriter(conn)
	for {
		bytes := <-writeChann
		if bytes == nil {
			return
		}
		w.WriteBytes(bytes)
	}
}
func (s *Server) closeConnection(conn net.Conn) {
	s.connWriteChans.Delete(conn)
	conn.Close()
}
