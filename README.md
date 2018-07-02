# Qpush, message push system writen in go


-------------------

## Try
```
go run app/server/main.go localhost:8888 localhost:8890 --env dev               #启动server

go run app/client/consumer/client.go localhost:8888 1008 ddddddd xuzhiqiang     #启动client

go run app/client/agent/agent.go push  localhost:8890 1 title content           #启动agent
```

