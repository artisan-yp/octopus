package xlog

import (
	"runtime"
	"strings"
	"time"

	"github.com/k8s-practice/octopus/xlog/internal/bufferpool"
)

type Entry struct {
	Level   Level
	Time    time.Time
	Message string
	Caller  Caller
}

type Caller struct {
	Defined  bool
	PC       uintptr
	File     string
	Line     int
	Function string
}

func (c *Caller) String() string {
	return c.FullPath()
}

func (c *Caller) FullPath() string {
	if !c.Defined {
		return ""
	}

	buf := bufferpool.Get()
	buf.AppendString(c.File)
	buf.AppendByte(':')
	buf.AppendInt(int64(c.Line))

	caller := buf.String()
	buf.Free()

	return caller
}

func (c *Caller) TrimedPath() string {
	if !c.Defined {
		return ""
	}

	idx := strings.LastIndexByte(c.File, '/')
	if idx == -1 {
		return c.FullPath()
	}

	idx = strings.LastIndexByte(c.File[:idx], '/')
	if idx == -1 {
		return c.FullPath()
	}

	buf := bufferpool.Get()
	buf.AppendString(c.File[idx+1:])
	buf.AppendByte(':')
	buf.AppendInt(int64(c.Line))

	caller := buf.String()
	buf.Free()

	return caller
}

func getCallerFrame(skip int) (frame runtime.Frame, ok bool) {
	const skipOffset = 2

	pc := make([]uintptr, 1)
	numFrames := runtime.Callers(skip+skipOffset, pc)
	if numFrames < 1 {
		return
	}

	frame, _ = runtime.CallersFrames(pc).Next()
	return frame, frame.PC != 0
}
