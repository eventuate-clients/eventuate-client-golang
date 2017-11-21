package logger

import (
	"log"
	"os"
)

type Logger interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

type LogLevelEnum int

const (
	Silent LogLevelEnum = iota
	Verbose
)

type nilLogger struct{}

func NewNilLogger() *nilLogger {
	return &nilLogger{}
}

func (*nilLogger) Println(v ...interface{})               {}
func (*nilLogger) Printf(format string, v ...interface{}) {}

func NewLogger(level LogLevelEnum) Logger {
	switch level {
	case Silent:
		{
			return NewNilLogger()
		}
	case Verbose:
		fallthrough
	default:
		{
			//return log.New(os.Stdout, "#LL# ", 0)
			return log.New(os.Stderr, "#EV#: ", log.LstdFlags)
		}

	}
}
