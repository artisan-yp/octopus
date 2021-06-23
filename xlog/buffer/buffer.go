package buffer

import (
	"strconv"
	"time"
)

const _size = 1024

type Buffer struct {
	bytes []byte
	pool  Pool
}

func (b *Buffer) AppendByte(v byte) {
	b.bytes = append(b.bytes, v)
}

func (b *Buffer) AppendString(s string) {
	b.bytes = append(b.bytes, s...)
}

func (b *Buffer) AppendInt(i int64) {
	b.bytes = strconv.AppendInt(b.bytes, i, 10)
}

func (b *Buffer) AppendTime(t time.Time, layout string) {
	b.bytes = t.AppendFormat(b.bytes, layout)
}

func (b *Buffer) AppendUint(i uint64) {
	b.bytes = strconv.AppendUint(b.bytes, i, 10)
}

func (b *Buffer) AppendBool(v bool) {
	b.bytes = strconv.AppendBool(b.bytes, v)
}

func (b *Buffer) AppendFloat(f float64, bitSize int) {
	b.bytes = strconv.AppendFloat(b.bytes, f, 'f', -1, bitSize)
}

func (b *Buffer) Len() int {
	return len(b.bytes)
}

func (b *Buffer) Cap() int {
	return cap(b.bytes)
}

func (b *Buffer) Bytes() []byte {
	return b.bytes
}

func (b *Buffer) String() string {
	return string(b.bytes)
}

func (b *Buffer) Reset() {
	b.bytes = b.bytes[:0]
}

func (b *Buffer) Write(bytes []byte) (int, error) {
	b.bytes = append(b.bytes, bytes...)
	return len(bytes), nil
}

func (b *Buffer) TrimNewline() {
	if i := len(b.bytes) - 1; i >= 0 {
		if b.bytes[i] == '\n' {
			b.bytes = b.bytes[:i]
		}
	}
}

func (b *Buffer) Free() {
	b.pool.put(b)
}
