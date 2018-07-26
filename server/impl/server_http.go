package impl

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"qpush/client"
	"qpush/modules/config"
	"qpush/modules/logger"
	"qpush/server"
	"runtime"
)

// this file is for debug only

func (s *Server) handleHTTP() {

	server := s.startHTTPServer()

	<-s.done

	if err := server.Shutdown(nil); err != nil {
		panic(err)
	}
}

func (s *Server) startHTTPServer() *http.Server {
	srv := &http.Server{Addr: "0.0.0.0:8080"}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if config.Get().Env == config.ProdEnv {
			return
		}

		if r.URL.Path != "/" {
			return
		}

		id := "1"
		title := "test title"
		content := "test content"

		msg := client.Msg{
			MsgID: id, Title: title, Content: content}
		payload, _ := json.Marshal(msg)
		packet := server.MakePacket(0, server.ForwardCmd, payload)

		s.Walk(func(conn net.Conn, ctx *server.ConnectionCtx) bool {

			if ctx.Internal {
				return true
			}

			select {
			case ctx.WriteChan <- packet:
			default:
				logger.Error("writeChan blocked for", ctx.GUID)
			}

			return true
		})

		runtime.GC()
		io.WriteString(w, "ok\n")
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("http ListenAndServe failed", err)
		}
	}()

	return srv
}
