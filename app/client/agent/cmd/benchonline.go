package cmd

import (
	"io/ioutil"
	"qpush/client"
	"qpush/server"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/zhiqiangxu/qrpc"
)

var benchOnlineCmd = &cobra.Command{
	Use:   "benchonline [internal address] [offset] [size]",
	Short: "connect to [internal address] and login",
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		internalAddress := args[0]
		offset, _ := strconv.Atoi(args[1])
		size, _ := strconv.Atoi(args[2])

		appID := 1001
		appKey := "UwMTA1Nw"

		bytes, err := ioutil.ReadFile("app/client/agent/cmd/guid.txt")
		if err != nil {
			panic(err)
		}
		guids := strings.Split(string(bytes), "\n")

		var conns []*client.Connection
		var wg sync.WaitGroup
		for i := 0; i < size; i++ {
			guid := strings.TrimSpace(guids[offset+i])
			qrpc.GoFunc(&wg, func() {
				conn, _ := client.NewConnection(internalAddress, qrpc.ConnectionConfig{}, nil)
				loginCmd := client.LoginCmd{AppID: appID, AppKey: appKey, GUID: guid}
				_, _, err := conn.Request(server.LoginCmd, 0, loginCmd)
				if err != nil {
					panic(err)
				}
				conns = append(conns, conn)
				// fmt.Println(string(frame.Payload))
			})
			if i%10 == 0 {
				wg.Wait()
			}
		}
		wg.Wait()

		select {}
	}}

func init() {
	rootCmd.AddCommand(benchOnlineCmd)
}
