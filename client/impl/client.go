package impl

import (
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
func (cli *Client) Dial(address string, guid string) *MsgConnection {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to dial %s", address))
		return nil
	}

	mc := &MsgConnection{conn: conn}
	cmd := &client.LoginCmd{GUID: guid}
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

	var requestID uint64
	packet := simpl.MakePacket(requestID, cmd, jsonBytes)

	w := simpl.NewStreamWriter(mc.conn)
	err = w.WriteBytes(packet)

	return requestID, err
}

// Subscribe messages from the underlying connection
func (mc *MsgConnection) Subscribe(cb *OnResponse) error {
	r := simpl.NewStreamReader(mc.conn)

	for {
		size, err := r.ReadUint32()
		logger.Debug(fmt.Sprintf("size is %d", size))
		if err != nil {
			return err
		}

		logger.Debug("test1")
		payload := make([]byte, size)
		err = r.ReadBytes(payload)
		if err != nil {
			return err
		}
		logger.Debug("test2")

		requestID := binary.BigEndian.Uint64(payload)
		cmd := server.Cmd(binary.BigEndian.Uint32(payload[8:]))

		ok := cb.Call(requestID, cmd, payload[8:])
		if !ok {
			return nil
		}
	}
}
