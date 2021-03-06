package tail

import (
	"context"
	"os"
	"qpush/pkg/logger"

	"github.com/gorilla/websocket"
	"github.com/hpcloud/tail"
)

// Push2WS start transfer tail output to ws
// websocket.Conn will be Closed after return
func Push2WS(c *websocket.Conn, file string, n int) error {

	defer c.Close()

	tailOffset, err := Line2Offset(file, n)
	// logger.Debug("tailOffset", tailOffset)
	if err != nil {
		logger.Error("Line2Offset", err)
		return err
	}

	location := &tail.SeekInfo{Offset: tailOffset, Whence: os.SEEK_SET}
	t, err := tail.TailFile(file, tail.Config{Follow: true, MustExist: true, Location: location})
	if t != nil {
		defer t.Stop()
	}
	if err != nil {
		err = c.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		if err != nil {
			logger.Error("WriteMessage:", err)
		}
		return err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go func() {
		for {
			_, _, err := c.ReadMessage()
			if _, ok := err.(*websocket.CloseError); ok {
				cancelFunc()
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	for {
		select {
		case line := <-t.Lines:
			if line == nil {
				return nil
			}
			err = c.WriteMessage(websocket.TextMessage, []byte(line.Text))
			if err != nil {
				logger.Error("WriteMessage:", err)
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}

	}
}
