package test

import (
	"encoding/json"
	"fmt"
	"qpush/client"
	cimpl "qpush/client/impl"
	"qpush/server"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	NumberConn      = 140
	NumberMesg      = 1000
	PublicAddress   = "106.14.50.182:8888"
	InternalAddress = "106.14.50.182:8890"
)

func TestMassiveConnections(t *testing.T) {
	t.Parallel()

	c := cimpl.NewClient()
	// make 10000 connections
	wgConn := sync.WaitGroup{}
	wgDone := sync.WaitGroup{}
	n := 0
	for n < NumberConn {

		wgConn.Add(1)
		wgDone.Add(1)
		go func(n int) {
			defer wgDone.Done()

			conn := c.Dial(PublicAddress, strconv.Itoa(n))
			if conn == nil {
				t.Fatalf("failed to dial")
				return
			}

			nextIdx := 1
			firstTime := true
			cb := cimpl.NewCallBack(func(requestID uint64, cmd server.Cmd, bytes []byte) bool {
				fmt.Println("cmd is", cmd)
				if firstTime {
					wgConn.Done()
					firstTime = false
				}

				if cmd == server.ForwardCmd {
					message := &client.PushCmd{}
					err := json.Unmarshal(bytes, message)
					if err != nil {
						t.Fatalf("failed to Unmarshal: %v, %v", err, string(bytes))
					}
					if message.MsgID != nextIdx {
						t.Fatalf("unexpected MsgID: %d vs %d", message.MsgID, nextIdx)
					}

					nextIdx++

					if nextIdx > NumberMesg {
						return false
					}
				}

				return true
			})

			go func() {
				for {
					time.Sleep(time.Second * 20)
					conn.SendCmd(server.HeartBeatCmd, nil)
				}
			}()
			err := conn.Subscribe(cb)
			if err != nil {
				t.Fatalf("Subscribe failed: %v", err)
			}

		}(n)

		n++
	}
	wgConn.Wait()

	// send 1000 messages
	agent := cimpl.NewAgent()
	conn := agent.Dial(InternalAddress)
	if conn == nil {
		t.Fatalf("failed to dial")
		return
	}

	idx := 1
	for idx <= NumberMesg {
		pushCmd := &client.PushCmd{MsgID: idx, Title: fmt.Sprintf("title %d", idx), Content: fmt.Sprintf("body %d", idx)}

		t.Log("pushing message\n")
		bytes, err := conn.SendCmdBlocking(server.PushCmd, pushCmd)
		if err != nil {
			t.Fatalf("SendCmdBlocking failed: %v", err)
		}

		fmt.Println(fmt.Sprintf("%d sent", idx), string(bytes))

		idx++
	}

	wgDone.Wait()

}
