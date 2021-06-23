package bufferpool

import "github.com/k8s-practice/octopus/xlog/buffer"

var (
	_pool = buffer.NewPool()
	Get   = _pool.Get
)
