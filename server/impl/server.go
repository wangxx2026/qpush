package impl

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"qpush/modules/logger"
	"qpush/server"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// Server is data struct for server
type Server struct {
	connWriteChans sync.Map
	guidConn       sync.Map // [string]net.Conn
	connCtx        sync.Map // [net.Conn]*ConnectionCtx
	readBufferSize int
	handler        server.Handler
	upTime         time.Time

	// for heartbeat
	hbConfig server.HeartBeatConfig
}

const (
	// DefaultAcceptTimeout is the default accept timeout duration
	DefaultAcceptTimeout = 5 * time.Second
	// DefaultReadTimeout is the default read timeout duration in seconds
	DefaultReadTimeout = 10 * 60 //TODO change back
	// DefaultInternalReadTimeout is read timeout for internal connections
	DefaultInternalReadTimeout = 10 * 60
	// DefaultWriteTimeout is default timeout for write
	DefaultWriteTimeout = time.Second * 10
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
		handler:        handler,
		upTime:         time.Now(),
		hbConfig:       c.HBConfig}
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

	wg.Add(1)
	go s.heartBeat(done, &wg)

	wg.Add(1)
	go s.handleHTTP(done, &wg)

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

// GetCtx return the ctx for connection
func (s *Server) GetCtx(conn net.Conn) *server.ConnectionCtx {
	ctx, ok := s.connCtx.Load(conn)
	if !ok {
		return nil
	}

	return ctx.(*server.ConnectionCtx)
}

// GetStatus get the server status
func (s *Server) GetStatus() *server.Status {

	guidCount := 0
	count := 0
	s.connWriteChans.Range(func(k, v interface{}) bool {

		count++

		ctx := s.GetCtx(k.(net.Conn))
		if ctx.GUID != "" {
			guidCount++
		}
		return true
	})

	var (
		GUIDConnMapSize int
		ConnCtxMapSize  int
	)

	s.guidConn.Range(func(k, v interface{}) bool {
		GUIDConnMapSize++
		return true
	})
	s.connCtx.Range(func(k, v interface{}) bool {
		ConnCtxMapSize++
		return true
	})

	status := &server.Status{
		Uptime: s.upTime, ConnectionCount: count, GUIDCount: guidCount,
		GUIDConnMapSize: GUIDConnMapSize,
		ConnCtxMapSize:  ConnCtxMapSize}

	runtime.ReadMemStats(&status.MemStats)

	return status
}

// BindGUIDToConn as names
func (s *Server) BindGUIDToConn(guid string, conn net.Conn) {
	oldConn, ok := s.guidConn.Load(guid)
	if ok {
		// TODO handle error
		s.CloseConnection(oldConn.(net.Conn))
	}
	s.guidConn.Store(guid, conn)
}

// KillGUID kills specified connection, usually from another goroutine
func (s *Server) KillGUID(guid string) error {
	conn, ok := s.guidConn.Load(guid)
	if !ok {
		return fmt.Errorf("no connection for guid:%s", guid)
	}

	return s.CloseConnection(conn.(net.Conn))
}

// MakePacket generate packet
func MakePacket(requestID uint64, cmd server.Cmd, payload []byte) []byte {
	length := 12 + uint32(len(payload))
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf, length)
	binary.BigEndian.PutUint64(buf[4:], requestID)
	binary.BigEndian.PutUint32(buf[12:], uint32(cmd))
	copy(buf[16:], payload)
	return buf
}

func (s *Server) handleSignal(quitChan chan os.Signal, done chan bool) {
	<-quitChan
	logger.Debug("signal captured")
	close(done)
}

func (s *Server) heartBeat(done chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(s.hbConfig.Interval)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			for {
				err := s.hbConfig.Callback()
				if err == nil {
					break
				}
				time.Sleep(time.Second)
			}
		}
	}

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

			// DEBUG only
			status := s.GetStatus()
			logger.Error("connection count is", status.ConnectionCount)

			continue
		}

		wg.Add(1)
		go s.handleConnection(conn, internal, done, wg)
	}
}

func (s *Server) handleConnection(conn net.Conn, internal bool, done chan bool, wg *sync.WaitGroup) {

	defer s.CloseConnection(conn)
	defer wg.Done()

	ctx := &server.ConnectionCtx{Internal: internal}
	//for query from other G
	s.connCtx.Store(conn, ctx)
	writeChan := make(chan []byte, 30)
	s.connWriteChans.Store(conn, writeChan)

	var readTimeout int
	if internal {
		readTimeout = DefaultInternalReadTimeout
	} else {
		readTimeout = DefaultReadTimeout
	}
	r := NewStreamReaderWithTimeout(conn, readTimeout)

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

			// 健康检查导致太多关闭，所以不输出日志
			//logger.Error(fmt.Sprintf("ReadUint32 failed:%s", err))

			return
		}

		if size < 12 {
			logger.Error("invalid packet", size)
			return
		}

		payload := make([]byte, size)
		err = r.ReadBytes(payload)
		if err != nil {

			// 健康检查导致太多关闭，所以不输出日志
			//logger.Error(fmt.Sprintf("ReadBytes failed:%s", err))

			return
		}

		requestID := binary.BigEndian.Uint64(payload)
		cmd := server.Cmd(binary.BigEndian.Uint32(payload[8:]))

		params := server.CmdParam{
			Param:     payload[12:],
			Server:    s,
			Conn:      conn,
			RequestID: requestID,
			Ctx:       ctx,
			Reader:    r}

		logger.Debug("cmdParam is", string(payload[12:]))
		responseCmd, response, err := s.handler.Call(cmd, internal, &params)
		if err != nil {
			logger.Error("handler.Call return error:%s", err)
			return
		}

		var (
			jsonResponse []byte
		)
		if response != nil {
			jsonResponse, err = json.Marshal(response)
			if err != nil {
				logger.Error("json.Marshal fail:%s", err)
				return
			}
		}

		logger.Debug("requestID", requestID, "responseCmd", responseCmd, "jsonResponse", string(jsonResponse))
		packet := MakePacket(requestID, responseCmd, jsonResponse)
		logger.Debug("packet is", packet)
		writeChan <- packet

	}

}

func (s *Server) handleWrite(conn net.Conn, writeChann chan []byte) {
	w := NewStreamWriter(conn)
	for {
		bytes := <-writeChann

		// logger.Debug("writeChann fired")
		if bytes == nil {
			return
		}

		// logger.Debug("WriteBytes called", bytes)
		err := w.WriteBytes(bytes)
		if err != nil {
			s.CloseConnection(conn)
			return
		}
	}
}

func (s *Server) CloseConnection(conn net.Conn) error {
	err := conn.Close()
	if err != nil {
		logger.Error("failed to close connection", err)
		return err
	}
	s.connWriteChans.Delete(conn)
	ctx, ok := s.connCtx.Load(conn)
	if ok && ctx.(*server.ConnectionCtx).GUID != "" {
		s.guidConn.Delete(ctx.(*server.ConnectionCtx).GUID)
	}
	s.connCtx.Delete(conn)
	return nil
}
