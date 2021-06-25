package xlog

import "os"

func init() {
}

type Log interface {
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Debugln(v ...interface{})

	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Infoln(v ...interface{})

	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Errorln(v ...interface{})

	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})

	SetLevel(lvl Level)
}

var std = NewStdLog(os.Stderr, "", DebugLevel, LstdFlags)

// Default returns the standard logger used by the package-level output functions.
func Default() *Logger { return std }
