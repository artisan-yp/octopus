package xlog

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/k8s-practice/octopus/xlog/rotatefile"
)

var cache = make(map[string]DepthLogger)

type severity int32

const (
	infoLog severity = iota
	warningLog
	errorLog
	fatalLog
	severityNum = fatalLog + 1
)

const severityChar = "IWEF"

var severityName = []string{
	infoLog:    "INFO",
	warningLog: "WARNING",
	errorLog:   "ERROR",
	fatalLog:   "FATAL",
}

type Level int32
type flushSyncWriter interface {
	io.Writer
	Flush() error
	Sync() error
}

type logger struct {
	biz string

	fileMutex sync.Mutex
	file      [severityNum]flushSyncWriter

	// log level.
	verbosity Level

	// hold buffer
	bufPool sync.Pool

	stop chan struct{}

	alsoToStderr bool
}

const d = 1

func (l *logger) WithAlsoToStderr(b bool) {
	l.alsoToStderr = b
}

func (l *logger) InfoDepth(depth int, args ...interface{}) {
	l.printDepth(infoLog, depth, args...)
}

func (l *logger) WarningDepth(depth int, args ...interface{}) {
	l.printDepth(warningLog, depth, args...)
}

func (l *logger) ErrorDepth(depth int, args ...interface{}) {
	l.printDepth(errorLog, depth, args...)
}

func (l *logger) FatalDepth(depth int, args ...interface{}) {
	l.printDepth(fatalLog, depth, args...)
}

func (l *logger) Info(args ...interface{}) {
	l.InfoDepth(d, args...)
}

func (l *logger) Infoln(args ...interface{}) {
	l.InfoDepth(d, fmt.Sprintln(args...))
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.InfoDepth(d, fmt.Sprintf(format, args...))
}

func (l *logger) Warning(args ...interface{}) {
	l.WarningDepth(d, args...)
}

func (l *logger) Warningln(args ...interface{}) {
	l.WarningDepth(d, fmt.Sprintln(args...))
}

func (l *logger) Warningf(format string, args ...interface{}) {
	l.WarningDepth(d, fmt.Sprintf(format, args...))
}

func (l *logger) Error(args ...interface{}) {
	l.ErrorDepth(d, args...)
}

func (l *logger) Errorln(args ...interface{}) {
	l.ErrorDepth(d, fmt.Sprintln(args...))
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.ErrorDepth(d, fmt.Sprintf(format, args...))
}

func (l *logger) Fatal(args ...interface{}) {
	l.FatalDepth(d, args...)
}

func (l *logger) Fatalln(args ...interface{}) {
	l.FatalDepth(d, fmt.Sprintln(args...))
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.FatalDepth(d, fmt.Sprintf(format, args...))
}

func (l *logger) V(lvl int) bool { return true }

func (l *logger) Sync() error {
	l.fileMutex.Lock()
	defer l.fileMutex.Unlock()

	l.flush()
	return nil
}

func (l *logger) output(s severity, buf *buffer, file string, line int, alsoToStderr bool) {
	l.fileMutex.Lock()

	data := buf.Bytes()
	if alsoToStderr {
		os.Stderr.Write(data)
	}

	if l.file[s] == nil {
		if err := l.createFiles(s); err != nil {
			os.Stderr.Write(data)
			l.exit(err)
		}
	}

	switch s {
	case fatalLog:
		l.file[fatalLog].Write(data)
		fallthrough
	case errorLog:
		l.file[errorLog].Write(data)
		fallthrough
	case warningLog:
		l.file[warningLog].Write(data)
		fallthrough
	case infoLog:
		l.file[infoLog].Write(data)
	}

	if s == fatalLog {
		l.flush()
		os.Exit(255)
	}

	l.fileMutex.Unlock()

	if buf.Len() < 256 {
		l.bufPool.Put(buf)
	}
}

func (l *logger) println(s severity, args ...interface{}) {
	buf, file, line := l.header(s, 0)
	fmt.Fprintln(buf, args...)
	l.output(s, buf, file, line, l.alsoToStderr)
}

func (l *logger) print(s severity, args ...interface{}) {
	l.printDepth(s, 1, args...)
}

func (l *logger) printf(s severity, format string, args ...interface{}) {
	buf, file, line := l.header(s, 0)
	fmt.Fprintf(buf, format, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	l.output(s, buf, file, line, l.alsoToStderr)
}

func (l *logger) printDepth(s severity, depth int, args ...interface{}) {
	buf, file, line := l.header(s, depth)
	fmt.Fprint(buf, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	l.output(s, buf, file, line, l.alsoToStderr)
}

func (l *logger) header(s severity, depth int) (*buffer, string, int) {
	fc := "???"
	pc, file, line, ok := runtime.Caller(3 + depth)
	if !ok {
		file = "???"
		line = 0
	} else {
		slash := strings.LastIndex(file, "/")
		if slash > 0 {
			file = file[slash+1:]
		}

		if Func := runtime.FuncForPC(pc); Func != nil {
			fc = Func.Name()
		}
	}

	return l.formatHeader(s, file, line, fc), file, line
}

func (l *logger) formatHeader(s severity, file string, line int, fc string) *buffer {
	if line < 0 {
		line = 0
	}
	if s > fatalLog {
		s = infoLog
	}
	buf := l.bufPool.Get().(*buffer)
	buf.Buffer.Reset()

	now := time.Now()
	_, month, day := now.Date()
	hour, minute, second := now.Clock()
	buf.tmp[0] = severityChar[s]
	buf.twoDigits(1, int(month))
	buf.twoDigits(3, day)
	buf.tmp[5] = '-'
	buf.twoDigits(6, hour)
	buf.tmp[8] = ':'
	buf.twoDigits(9, minute)
	buf.tmp[11] = ':'
	buf.twoDigits(12, second)
	buf.tmp[14] = '.'
	buf.nDigits(6, 15, now.Nanosecond()/1000, '0')
	buf.tmp[21] = ' '
	// TODO: should be TID
	buf.nDigits(7, 22, os.Getpid(), ' ')
	buf.tmp[29] = ' '

	buf.Write(buf.tmp[:30])
	buf.WriteString(file)

	buf.tmp[0] = ':'
	n := buf.someDigits(1, line)
	buf.tmp[n+1] = ':'
	buf.Write(buf.tmp[:n+2])

	buf.WriteString("[" + fc + "] ")

	return buf
}

func (l *logger) createFiles(sev severity) error {
	for s := sev; s >= infoLog && l.file[s] == nil; s-- {
		rf, err := rotatefile.New(l.biz, severityName[s])
		if err != nil {
			return err
		} else {
			l.file[s] = rf
		}
	}

	return nil
}

func (l *logger) flush() {
	for s := infoLog; s <= fatalLog && l.file[s] != nil; s++ {
		l.file[s].Flush()
		l.file[s].Sync()
	}
}

func (l *logger) exit(err error) {
	fmt.Fprintf(os.Stderr, "log: exiting because of error: %s\n", err)
	os.Exit(2)
}

func Component(name string) DepthLogger {
	if logger, ok := cache[name]; ok {
		return logger
	}

	l := &logger{
		biz: name,
		bufPool: sync.Pool{
			New: func() interface{} { return new(buffer) },
		},
		stop: make(chan struct{}, 0),
	}
	cache[name] = l

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				l.Sync()
			case <-l.stop:
				l.Sync()
				break
			}
		}
	}()

	return l
}
