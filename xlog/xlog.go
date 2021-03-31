package xlog

type Logger interface {
	Info(args ...interface{})
	Infoln(args ...interface{})
	Infof(format string, args ...interface{})

	Warning(args ...interface{})
	Warningln(args ...interface{})
	Warningf(format string, args ...interface{})

	Error(args ...interface{})
	Errorln(args ...interface{})
	Errorf(format string, args ...interface{})

	Fatal(args ...interface{})
	Fatalln(args ...interface{})
	Fatalf(format string, args ...interface{})

	// V reports whether verbosity level l is at least the requested verbose level.
	V(l int) bool

	// sync buffer content.
	Sync() error
}

type DepthLogger interface {
	Logger

	InfoDepth(depth int, args ...interface{})
	WarningDepth(depth int, args ...interface{})
	ErrorDepth(depth int, args ...interface{})
	FatalDepth(depth int, args ...interface{})
}
