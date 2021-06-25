package xlog

import (
	"io"
	"sync"
)

type lockedWriter struct {
	sync.Mutex
	writer io.Writer
}

func (lw *lockedWriter) Write(p []byte) (int, error) {
	lw.Lock()
	defer lw.Unlock()
	return lw.writer.Write(p)
}
