package buffer

import "sync"

type Pool struct {
	pool *sync.Pool
}

func NewPool() Pool {
	return Pool{
		pool: &sync.Pool{
			New: func() interface{} {
				return &Buffer{bytes: make([]byte, 0, _size)}
			},
		},
	}
}

func (p Pool) Get() *Buffer {
	buf := p.pool.Get().(*Buffer)
	buf.Reset()
	buf.pool = p
	return buf
}

func (p Pool) put(buff *Buffer) {
	p.pool.Put(buff)
}
