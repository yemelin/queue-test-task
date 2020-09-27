package main

import (
	"fmt"
	"log"
	"os"
)

var logLevel = os.Getenv("LOGLEVEL")

type Logger struct {
	*log.Logger
	debug bool
}

func NewLogger(name string) *Logger {
	l := log.New(os.Stdout, prefix(name), log.Lmsgprefix|log.Lmicroseconds)
	return &Logger{l, logLevel == "DEBUG"}
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.debug {
		l.Printf(format, v...)
	}
}

func prefix(s string) string {
	return fmt.Sprintf("%-15s", fmt.Sprintf("[%s]:", s))
}
