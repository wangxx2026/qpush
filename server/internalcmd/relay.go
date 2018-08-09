package internalcmd

import (
	"qpush/pkg/logger"
	"sync"

	"github.com/zhiqiangxu/qrpc"
)

const (
	// SessionStartCmd when pair matched
	SessionStartCmd qrpc.Cmd = iota
	// MsgCmd for message
	MsgCmd
)

// RelayCmd do relay
type RelayCmd struct {
	mu             sync.Mutex
	first2second   chan string
	second2firstch chan chan string
}

// ServeQRPC implements qrpc.Handler
func (cmd *RelayCmd) ServeQRPC(writer qrpc.FrameWriter, frame *qrpc.RequestFrame) {

	cmd.mu.Lock()
	first2second := cmd.first2second
	if first2second == nil {
		// the first to join
		first2second = make(chan string)
		second2firstch := make(chan chan string, 1)
		cmd.first2second = first2second
		cmd.second2firstch = second2firstch
		cmd.mu.Unlock()

		select {
		case second2first := <-second2firstch:
			cmd.relayMsg(writer, frame, second2first, first2second)
		case <-frame.Context().Done():
			cmd.mu.Lock()
			if cmd.first2second == first2second {
				cmd.first2second = nil
				cmd.second2firstch = nil
			}
			cmd.mu.Unlock()
			return
		}

	} else {
		// the second to join
		second2first := make(chan string)
		second2firstch := cmd.second2firstch
		cmd.first2second = nil
		cmd.second2firstch = nil
		cmd.mu.Unlock()

		second2firstch <- second2first

		cmd.relayMsg(writer, frame, first2second, second2first)
	}

}

func (cmd *RelayCmd) relayMsg(writer qrpc.FrameWriter, frame *qrpc.RequestFrame, p2s chan string, s2p chan string) {

	defer close(s2p)

	writer.StartWrite(frame.RequestID, SessionStartCmd, qrpc.StreamFlag)
	err := writer.EndWrite()
	if err != nil {
		return
	}

	for {
		select {
		case <-frame.Context().Done():
			return
		case f := <-frame.FrameCh():
			logger.Debug("test1")
			if f == nil {
				return
			}
			logger.Debug("test2", "payload", string(f.Payload))

		L:
			for {
				select {
				case s2p <- string(f.Payload):
					logger.Debug("test3")
					break L
				case msg := <-p2s:
					if msg == "" {
						logger.Debug("test5")
						return
					}
					if cmd.writeMsg(writer, frame, msg) != nil {
						return
					}
				}
			}
			logger.Debug("test4")

		case msg := <-p2s:
			if msg == "" {
				logger.Debug("test6")
				return
			}
			if cmd.writeMsg(writer, frame, msg) != nil {
				return
			}
		}
	}
}

func (cmd *RelayCmd) writeMsg(writer qrpc.FrameWriter, frame *qrpc.RequestFrame, msg string) error {
	writer.StartWrite(frame.RequestID, MsgCmd, qrpc.StreamFlag)
	writer.WriteBytes([]byte(msg))
	return writer.EndWrite()
}
