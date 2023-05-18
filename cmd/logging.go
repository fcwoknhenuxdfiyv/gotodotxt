package cmd

import (
	"io"
	"os"
	"path"

	"github.com/op/go-logging"
)

// Log levels.
const (
	CRITICAL int = iota
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

type logFuncSig func(args ...interface{})
type logfFuncSig func(format string, args ...interface{})

func GetLogger(level int, foreground bool) *logging.Logger {
	if cfgDir == "" {
		panic("No cfgDir")
	}
	// fmt.Println(filename)
	l := logging.MustGetLogger("main")
	l.ExtraCalldepth = 1
	logFormat := logging.MustStringFormatter(
		`%{color}%{time:06-01-02 15:04:05} %{level:.1s} ` +
			`%{shortfile:20s} %{color:reset}%{message}`,
	)
	var logfile io.Writer
	var err error
	if !foreground {
		logfile, err = os.OpenFile(path.Join(cfgDir, "gotodotxt.log"),
			os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			panic("error opening log file: " + err.Error())
		}
	} else {
		logfile = os.Stdout
	}
	backend := logging.NewLogBackend(logfile, "", 0)
	formatter := logging.NewBackendFormatter(backend, logFormat)
	leveled := logging.AddModuleLevel(formatter)
	leveled.SetLevel(logging.Level(level), "")
	logging.SetBackend(leveled)

	return l
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

// func log(level int, msg string) {
// 	if Log == nil || msg == "" {
// 		return
// 	}
// 	logf(level, "%s", msg)
// }
//
// func logf(level int, format string, args ...any) {
// 	if Log == nil || len(args) == 0 {
// 		return
// 	}
// 	switch {
// 	case level == CRITICAL:
// 		Log.Criticalf(format, args...)
// 	case level == ERROR:
// 		Log.Errorf(format, args...)
// 	case level == WARNING:
// 		Log.Warningf(format, args...)
// 	case level == NOTICE:
// 		Log.Noticef(format, args...)
// 	case level == INFO:
// 		Log.Infof(format, args...)
// 	case level == DEBUG:
// 		Log.Debugf(format, args...)
// 	}
// }
