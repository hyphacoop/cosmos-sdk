package pooledmessage

import "github.com/cosmos/gogoproto/proto"

type PooledMessage[T proto.Message] struct {
	Value T
	pool  *Pool[T]
}

func (p *PooledMessage[T]) Release() {
	p.pool.put(p.Value)
}
