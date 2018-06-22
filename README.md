# Qpush, message push system writen in go


-------------------

## Try
```
go run app/server/main.go localhost:8888 localhost:8890         #启动server

go run app/client/client.go "localhost:8888" "guid"             #启动client

go run app/client/agent.go push localhost:8890 1 title content3 #启动agent
```

