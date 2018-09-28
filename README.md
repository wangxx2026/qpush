# Qpush, message push system writen in go


-------------------

## Try
```bash
go run app/server/main.go localhost:8888 localhost:8890 --env dev               #启动server

go run app/client/consumer/client.go localhost:8888 1008 ddddddd cedce1d0-42b0-500e-93d7-0ab38dd105b9     #启动client

go run app/client/agent/agent.go push  localhost:8890 1008 1 title content           #启动agent
```

