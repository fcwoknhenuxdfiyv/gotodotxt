package tdt

import "github.com/op/go-logging"

// Log levels.
const (
	CRITICAL int = iota
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

var (
	log *logging.Logger
)

type logFuncSig func(args ...interface{})
type logfFuncSig func(format string, args ...interface{})

func SetLogger(l *logging.Logger) {
	log = l
}

func Log(f logFuncSig, args ...interface{}) {
	if log == nil {
		return
	}
	f(args...)
}

func Logf(f logfFuncSig, format string, args ...interface{}) {
	if log == nil {
		return
	}
	f(format, args...)
}
