package logger

import (
	"fmt"
	"os"
)

// Info prints to stdout
func Info(msg ...interface{}) {
	fmt.Println(msg...)
}

// Error prints to stderr
func Error(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, msg...)
}
