package cmd

import (
	"context"
	"io/ioutil"
	"qpush/client"
	"qpush/server"
	"strconv"
	"strings"

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

		api := client.NewAPI([]string{internalAddress}, qrpc.ConnectionConfig{}, nil)
		for i := 0; i < size; i++ {
			loginCmd := client.LoginCmd{AppID: appID, AppKey: appKey, GUID: strings.TrimSpace(guids[offset+i])}
			_, err := api.CallForFrame(context.Background(), server.LoginCmd, loginCmd)
			if err != nil {
				panic(err)
			}
			// fmt.Println(string(frame.Payload))
		}

		select {}
	}}

func init() {
	rootCmd.AddCommand(benchOnlineCmd)
}
