package logger

import (
	"fmt"
	"os"
	"qpush/modules/config"
)

// Info prints to stdout
func Info(msg ...interface{}) {
	fmt.Println(msg...)
}

// Error prints to stderr
func Error(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, msg...)
}

// Debug prints to stderr
func Debug(msg ...interface{}) {

	conf := config.Get()

	if conf.Env != config.DevEnv {
		return
	}

	fmt.Fprintln(os.Stderr, msg...)
}
