package logger

import (
	"fmt"
	"os"
	"qpush/modules/config"
	"time"
)

const (
	// SEP for log sperator
	SEP = " : "
)

// Info prints to stdout
func Info(msg ...interface{}) {
	fmt.Print(time.Now().String(), SEP)
	fmt.Println(msg...)
}

// Error prints to stderr
func Error(msg ...interface{}) {
	fmt.Fprint(os.Stderr, time.Now().String(), SEP)
	fmt.Fprintln(os.Stderr, msg...)
}

// Debug prints to stderr
func Debug(msg ...interface{}) {

	conf := config.Get()

	if conf.Env != config.DevEnv {
		return
	}

	fmt.Fprint(os.Stderr, time.Now().String(), SEP)
	fmt.Fprintln(os.Stderr, msg...)
}
