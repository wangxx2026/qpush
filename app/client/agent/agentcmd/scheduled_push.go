package agentcmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"qpush/modules/config"
	"qpush/modules/logger"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/spf13/cobra"
)

// this file implements scheduled push service

const (
	messageOffsetKey = "qpush-agent:message-offset"
)

var (
	env string
)

var scheduledPushCmd = &cobra.Command{
	Use:   "scheduled_push",
	Short: "get messages to send and push to server",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()

		for {

			resp, err := http.Get("http://baidu.com/")
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			logger.Debug(string(body))
			lastMsgID := getMessageOffset()

			setMessageOffset(lastMsgID)
		}
	}}

func getMessageOffset() int {
	client := getRedisClient()

	value, err := client.Get(messageOffsetKey).Result()
	if err == redis.Nil {
		return 0
	} else if err != nil {
		panic(err)
	} else {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			logger.Error("invalid message offset", value)
			intValue = 0
		}
		return intValue
	}
}

func getRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}

	return client
}

func setMessageOffset(msgID int) {
	client := getRedisClient()

	err := client.Set(messageOffsetKey, msgID, 0).Err()
	if err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.AddCommand(scheduledPushCmd)
	scheduledPushCmd.Flags().StringVar(&env, "env", "", "environment")
}

func initConfig() {
	_, err := config.Load(env)
	if err != nil {
		panic(fmt.Sprintf("failed to load config file: %s", env))
	}
}
