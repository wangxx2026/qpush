package impl

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"qpush/client"
	"qpush/modules/logger"
	"qpush/server"
	simpl "qpush/server/impl"
)

// Client is data structor for client
type Client struct {
}

// NewClient creates a client instance
func NewClient() *Client {
	return &Client{}
}

// Dial is called to initiate connections
func (cli *Client) Dial(address string, cmd *client.LoginCmd) *MsgConnection {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to dial %s", address))
		return nil
	}

	mc := &MsgConnection{conn: conn}
	_, err = mc.SendCmd(server.LoginCmd, cmd)
	if err != nil {
		logger.Error("failed to send login cmd", err)
		return nil
	}
	return mc
}

// MsgConnection the data struct for underlying connection
type MsgConnection struct {
	conn net.Conn
}

// SendCmd sends a cmd to server
func (mc *MsgConnection) SendCmd(cmd server.Cmd, cmdParam interface{}) (uint64, error) {
	var (
		jsonBytes []byte
		err       error
	)
	// cmdParam is nil for HearBeatCmd
	if cmdParam != nil {
		jsonBytes, err = json.Marshal(cmdParam)
		if err != nil {
			return 0, err
		}
	}

	requestID := PoorManUUID()
	packet := simpl.MakePacket(requestID, cmd, jsonBytes)

	w := simpl.NewStreamWriter(mc.conn)
	err = w.WriteBytes(packet)

	return requestID, err
}

// PoorManUUID generate a uint64 uuid
func PoorManUUID() uint64 {
	buf := make([]byte, 8)
	rand.Read(buf)
	return binary.LittleEndian.Uint64(buf)
}

// SendCmdBlocking will block indefinetely
func (mc *MsgConnection) SendCmdBlocking(cmd server.Cmd, cmdParam interface{}) ([]byte, error) {
	return mc.SendCmdBlockingWithTimeout(cmd, cmdParam, simpl.ReadNoTimeout)
}

// SendCmdBlockingWithTimeout works in blocking mode
func (mc *MsgConnection) SendCmdBlockingWithTimeout(cmd server.Cmd, cmdParam interface{}, seconds int) ([]byte, error) {
	ID, err := mc.SendCmd(cmd, cmdParam)
	if err != nil {
		logger.Error("failed to send cmd", cmd, cmdParam, err)
		return nil, err
	}

	var byteResp []byte
	cb := NewCallBack(func(requestID uint64, cmd server.Cmd, bytes []byte) bool {

		if ID == requestID {
			logger.Debug("got reply")
			logger.Debug(requestID, cmd, string(bytes))
			byteResp = make([]byte, len(bytes))
			copy(byteResp, bytes)
			return false
		}
		return true
	})

	err = mc.SubscribeWithTimeout(cb, seconds)

	if err != nil {
		logger.Error("Subscribe error", err)
		return nil, err
	}

	return byteResp, nil

}

// Subscribe messages from the underlying connection
func (mc *MsgConnection) Subscribe(cb *OnResponse) error {
	return mc.SubscribeWithTimeout(cb, simpl.ReadNoTimeout)
}

// SubscribeWithTimeout is Subscribe with timeout
func (mc *MsgConnection) SubscribeWithTimeout(cb *OnResponse, seconds int) error {
	r := simpl.NewStreamReaderWithTimeout(mc.conn, seconds)

	for {
		size, err := r.ReadUint32()
		logger.Debug(fmt.Sprintf("size is %d", size))
		if err != nil {
			return err
		}

		payload := make([]byte, size)
		err = r.ReadBytes(payload)
		if err != nil {
			return err
		}

		requestID := binary.BigEndian.Uint64(payload)
		cmd := server.Cmd(binary.BigEndian.Uint32(payload[8:]))

		ok := cb.Call(requestID, cmd, payload[12:])
		if !ok {
			return nil
		}
	}
}
