package pooledmessage

import (
	"sync"

	"github.com/cosmos/gogoproto/proto"
)

type Pool[T proto.Message] struct {
	pool *sync.Pool
}

func NewPool[T proto.Message](factory func() T) *Pool[T] {
	return &Pool[T]{
		pool: &sync.Pool{
			New: func() any {
				return factory()
			},
		},
	}
}

func (p *Pool[T]) Get() PooledMessage[T] {
	obj := p.pool.Get().(T)
	return PooledMessage[T]{
		Value: obj,
		pool:  p,
	}
}

func (p *Pool[T]) put(obj T) {
	p.pool.Put(obj)
}
