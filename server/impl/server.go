package impl

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"qpush/modules/logger"
	"qpush/modules/stream"
	"qpush/server"
	"qpush/server/impl/cmd"
	"qpush/server/impl/internalcmd"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
)

// Server is data struct for server
type Server struct {
	guidConn       sync.Map // [string]net.Conn
	connCtx        sync.Map // [net.Conn]*ConnectionCtx
	readBufferSize int
	handler        server.Handler
	upTime         time.Time
	routinesGroup  sync.WaitGroup
	done           chan bool

	// for heartbeat
	hbConfig server.HeartBeatConfig
}

const (
	// DefaultAcceptTimeout is the default accept timeout duration
	DefaultAcceptTimeout = 5 * time.Second
	// DefaultReadTimeout is the default read timeout duration in seconds
	DefaultReadTimeout = 10
	// DefaultInternalReadTimeout is read timeout for internal connections
	DefaultInternalReadTimeout = 10 * 60
	// DefaultWriteTimeout is default timeout for write in seconds
	DefaultWriteTimeout = 60
)

var (
	closedChan = make(chan bool)
)

func init() {
	close(closedChan)
}

// NewServer creates a server instance
func NewServer(c *server.Config) *Server {
	var (
		readBufferSize int
	)
	if c == nil {
		readBufferSize = server.DefaultReadBufferSize
	} else {
		readBufferSize = c.ReadBufferSize
	}

	serverHandler := &ServerHandler{}
	serverHandler.RegisterCmd(server.LoginCmd, false, &cmd.LoginCmd{})
	serverHandler.RegisterCmd(server.AckCmd, false, cmd.NewAckCmd())
	serverHandler.RegisterCmd(server.HeartBeatCmd, false, &cmd.HeartBeatCmd{})

	serverHandler.RegisterCmd(server.PushCmd, true, &internalcmd.PushCmd{})
	serverHandler.RegisterCmd(server.StatusCmd, true, &internalcmd.StatusCmd{})
	serverHandler.RegisterCmd(server.KillCmd, true, &internalcmd.KillCmd{})
	serverHandler.RegisterCmd(server.KillAllCmd, true, &internalcmd.KillAllCmd{})
	serverHandler.RegisterCmd(server.ListGUIDCmd, true, &internalcmd.ListGUIDCmd{})

	return &Server{
		readBufferSize: readBufferSize,
		handler:        serverHandler,
		done:           make(chan bool),
		upTime:         time.Now(),
		hbConfig:       c.HBConfig}
}

// ListenAndServe start listen and serve
func (s *Server) ListenAndServe(address string, internalAddress string) error {

	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	s.goFunc(func() {
		s.listenAndServe(address, false)
	})
	s.goFunc(func() {
		s.listenAndServe(internalAddress, true)
	})
	s.goFunc(func() {
		s.heartBeat()
	})
	s.goFunc(func() {
		s.handleHTTP()
	})
	s.goFunc(func() {
		s.handleSignal(quitChan)
	})

	s.waitShutdown()

	return nil
}

func (s *Server) goFunc(f func()) {
	s.routinesGroup.Add(1)
	go func() {
		defer s.routinesGroup.Done()
		f()
	}()
}

func (s *Server) waitShutdown() {
	s.routinesGroup.Wait()
}

// Walk walks each connection
func (s *Server) Walk(f func(net.Conn, *server.ConnectionCtx) bool) {
	s.connCtx.Range(func(k, v interface{}) bool {
		return f(k.(net.Conn), v.(*server.ConnectionCtx))
	})
}

// GetStatus get the server status
func (s *Server) GetStatus() *server.Status {

	guidCount := 0
	s.connCtx.Range(func(k, v interface{}) bool {

		ctx := v.(*server.ConnectionCtx)
		if ctx.GUID != "" {
			guidCount++
		}
		return true
	})

	var (
		GUIDConnMapSize int
		ConnCtxMapSize  int
	)

	handlerStatus := make(map[server.Cmd]interface{})
	s.handler.Walk(func(cmd server.Cmd, internal bool, cmdHandler server.CmdHandler) bool {

		cmdStatus, ok := cmdHandler.(server.StatusAble)
		if !ok {
			return true
		}
		status := cmdStatus.Status()
		if status != nil {
			handlerStatus[cmd] = status
		}

		return true
	})

	s.guidConn.Range(func(k, v interface{}) bool {
		GUIDConnMapSize++
		return true
	})
	s.connCtx.Range(func(k, v interface{}) bool {
		ConnCtxMapSize++
		return true
	})

	status := &server.Status{
		Uptime: s.upTime, GUIDCount: guidCount,
		GUIDConnMapSize: GUIDConnMapSize,
		ConnCtxMapSize:  ConnCtxMapSize,
		HandleStatus:    handlerStatus}

	return status
}

// BindAppGUIDToConn as names
func (s *Server) BindAppGUIDToConn(appid int, guid string, conn net.Conn) {
	appGUID := getAppGUID(appid, guid)

	oldConn, ok := s.guidConn.Load(appGUID)
	if ok {
		// TODO handle error
		closeChan, _ := s.CloseConnection(oldConn.(net.Conn))
		<-closeChan
	}

	s.guidConn.Store(appGUID, conn)
}

// SendTo send packet to specified connection
func (s *Server) SendTo(appid int, guids []string, packet []byte) int {
	var count int
	for _, guid := range guids {
		conn, ok := s.guidConn.Load(getAppGUID(appid, guid))
		if !ok {
			continue
		}

		ctx, ok := s.connCtx.Load(conn)
		if !ok {
			continue
		}

		select {
		case ctx.(*server.ConnectionCtx).WriteChan <- packet:
			count++
		default:
			continue
		}
	}

	return count
}

func getAppGUID(appID int, guid string) string {
	return fmt.Sprintf("%d:%s", appID, guid)
}

// KillAppGUID kills specified connection, usually from another goroutine
func (s *Server) KillAppGUID(appID int, guid string) error {
	appGUID := getAppGUID(appID, guid)

	conn, ok := s.guidConn.Load(appGUID)
	if !ok {
		return fmt.Errorf("no connection for guid:%s", appGUID)
	}

	_, err := s.CloseConnection(conn.(net.Conn))
	return err
}

func (s *Server) handleSignal(quitChan chan os.Signal) {
	<-quitChan
	logger.Debug("signal captured")
	close(s.done)
}

func (s *Server) heartBeat() {

	ticker := time.NewTicker(s.hbConfig.Interval)
	for {
		select {
		case <-s.done:
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

func (s *Server) listenAndServe(address string, internal bool) {

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
		case <-s.done:
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
			logger.Error("connection count is", status.ConnCtxMapSize)

			continue
		}

		go s.handleConnection(conn, internal)
	}
}

func (s *Server) handleConnection(conn net.Conn, internal bool) {

	defer func() {
		if err := recover(); err != nil {
			logger.Error("recovered from panic in handleConnection", err, string(debug.Stack()))
		}
	}()
	defer s.CloseConnection(conn)

	// store pointer
	ctx := &server.ConnectionCtx{
		Internal: internal, WriteChan: make(chan []byte, 30),
		CloseChan: make(chan bool)}
	// for query from other G
	s.connCtx.Store(conn, ctx)

	var readTimeout int
	if internal {
		readTimeout = DefaultInternalReadTimeout
	} else {
		readTimeout = DefaultReadTimeout
	}
	r := stream.NewReaderWithTimeout(conn, readTimeout)

	go s.handleWrite(conn, ctx.WriteChan, ctx.CloseChan)

	for {
		select {
		case <-s.done:
			return
		case <-ctx.CloseChan:
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
			logger.Error("handler.Call return error:", err)
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
		packet := server.MakePacket(requestID, responseCmd, jsonResponse)
		logger.Debug("packet is", packet)
		ctx.WriteChan <- packet

	}

}

func (s *Server) handleWrite(conn net.Conn, writeChann chan []byte, closeChan chan bool) {
	w := stream.NewWriterWithTimeout(conn, DefaultWriteTimeout)
	for {
		select {
		case <-s.done:
			return
		case <-closeChan:
			return
		case bytes := <-writeChann:

			// logger.Debug("writeChann fired")
			if bytes == nil {
				return
			}

			// logger.Debug("WriteBytes called", bytes)
			err := w.Write(bytes)
			if err != nil {
				s.CloseConnection(conn)
				return
			}
		}
	}
}

// CloseConnection close specified connection
// the return channel block until actually closed
func (s *Server) CloseConnection(conn net.Conn) (<-chan bool, error) {

	ctxInterface, ok := s.connCtx.Load(conn)
	if !ok {
		return closedChan, fmt.Errorf("conn already closed")
	}
	ctx := ctxInterface.(*server.ConnectionCtx)
	err := conn.Close()

	if err != nil {
		logger.Error("failed to close connection", err)
		return ctx.CloseChan, err
	}

	if ctx.GUID != "" {
		s.guidConn.Delete(
			getAppGUID(ctx.AppID, ctx.GUID))
	}

	// must keep this order, otherwise bug
	s.connCtx.Delete(conn)
	close(ctx.CloseChan)

	return ctx.CloseChan, nil
}
