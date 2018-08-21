package logger

import (
	"fmt"
	"os"
	"qpush/pkg/config"
	"time"

	"github.com/zhiqiangxu/qrpc"
)

const (
	// SEP for log sperator
	SEP = " : "
)

// Info prints to stdout
func Info(msg ...interface{}) {
	fmt.Fprint(os.Stdout, time.Now().String(), SEP, fmt.Sprintln(msg...))
}

// Error prints to stderr
func Error(msg ...interface{}) {
	now := time.Now().String()
	fmt.Fprint(os.Stdout, now, SEP, fmt.Sprintln(msg...))
	fmt.Fprint(os.Stderr, now, SEP, fmt.Sprintln(msg...))
}

// Debug prints to stderr
func Debug(msg ...interface{}) {

	conf := config.Get()

	if !conf.EnableDebug {
		return
	}

	Error(msg...)
}

type log struct {
}

func (l log) Info(msg ...interface{}) {
	Info(msg...)
}
func (l log) Error(msg ...interface{}) {
	Error(msg...)
}
func (l log) Debug(msg ...interface{}) {
	Debug(msg...)
}
func init() {
	qrpc.Logger = log{}
}
