package clipool

import (
	"sync"

	"google.golang.org/grpc"
)

type ClientPool struct {
	clients map[string]*grpc.ClientConn
	mu      sync.RWMutex
}

func New() *ClientPool {
	return &ClientPool{
		clients: make(map[string]*grpc.ClientConn),
	}
}

func (pool *ClientPool) Get(target string) (*grpc.ClientConn, error) {
	{
		pool.mu.RLock()
		defer pool.mu.Unlock()

		if client, ok := pool.clients[target]; ok {
			return client, nil
		}
	}

	pool.mu.Lock()
	defer pool.mu.Unlock()

	if client, err := grpc.Dial(target, grpc.WithInsecure()); err != nil {
		return nil, err
	} else {
		pool.clients[target] = client
		return client, nil
	}
}

func (pool *ClientPool) Put(target string) (*grpc.ClientConn, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if client, err := grpc.Dial(target, grpc.WithInsecure()); err != nil {
		return nil, err
	} else {
		pool.clients[target] = client
		return client, nil
	}
}
