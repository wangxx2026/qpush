package impl

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"qpush/client"
	"qpush/modules/logger"
	"qpush/server"
	"sync"
)

// this file is for debug only

func (s *Server) handleHTTP(done chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	server := s.startHTTPServer()

	<-done

	if err := server.Shutdown(nil); err != nil {
		panic(err)
	}
}

func (s *Server) startHTTPServer() *http.Server {
	srv := &http.Server{Addr: "0.0.0.0:8080"}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		id := 1
		title := "test title"
		content := "test content"

		cmd := &client.PushCmd{MsgID: id, Title: title, Content: content}
		payload, _ := json.Marshal(cmd)
		packet := MakePacket(0, server.ForwardCmd, payload)

		s.Walk(func(conn net.Conn, writeChan chan []byte) bool {

			ctx := s.GetCtx(conn)
			if ctx.Internal {
				return true
			}

			select {
			case writeChan <- packet:
			default:
				logger.Error("writeChan blocked for", ctx.GUID)
			}

			return true
		})

		io.WriteString(w, "ok\n")
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("http ListenAndServe failed", err)
		}
	}()

	return srv
}
