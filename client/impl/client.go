package impl

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"push-msg/modules/logger"
	simpl "push-msg/server/impl"
)

// Client is data structor for client
type Client struct {
}

// NewClient creates a client instance
func NewClient() *Client {
	return &Client{}
}

type loginCmd struct {
	Cmd  string `json:"cmd"`
	GUID string `json:"guid"`
}

// Dial is called to initiate connections
func (cli *Client) Dial(address string, guid string) *MsgConnection {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to dial %s", address))
		return nil
	}

	mc := &MsgConnection{conn: conn}
	cmd := &loginCmd{Cmd: "login", GUID: guid}
	jsonBytes, err := json.Marshal(cmd)
	logger.Info(string(jsonBytes))
	if err != nil {
		logger.Error(fmt.Sprintf("failed to marshal json:%s", err))
		return nil
	}
	_, err = mc.Send(jsonBytes)
	if err != nil {
		logger.Error("login failed")
		return nil
	}
	//TODO wait for reply
	return mc
}

// MsgConnection the data struct for underlying connection
type MsgConnection struct {
	conn net.Conn
}

// Send a message to the underlying connection
func (mc *MsgConnection) Send(jsonBytes []byte) (uint64, error) {
	var requestID uint64 = 0
	w := simpl.NewStreamWriter(mc.conn)
	w.WriteUint32(uint32(8 + len(jsonBytes)))
	w.WriteUint64(requestID)
	w.WriteBytes(jsonBytes)
	return requestID, nil
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

		payload := make([]byte, size)
		err = r.ReadBytes(payload)
		if err != nil {
			return err
		}

		requestID := binary.BigEndian.Uint64(payload)

		err = cb.Call(requestID, payload[8:])
		if err != nil {
			return err
		}
	}
}
