FROM h1.ywopt.com/library/golang:1.10.0-stretch

ENV EXEC_FILE app/server/main.go

EXPOSE 8888 8890 8080
RUN mkdir -p /go/src/qpush
ADD . /go/src/qpush
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && dpkg-reconfigure -f noninteractive tzdata

WORKDIR /go/src/qpush

ENTRYPOINT go run ${EXEC_FILE} 0.0.0.0:8888 0.0.0.0:8890 --env test