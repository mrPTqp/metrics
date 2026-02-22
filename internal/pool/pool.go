package pool

import "sync"

type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	sp sync.Pool
}

func NewPool[T Resetter]() *Pool[T] {
	return &Pool[T]{
		sp: sync.Pool{
			New: func() any {
				var zero T
				return zero
			},
		},
	}
}

func (p *Pool[T]) Get() T {
	v := p.sp.Get()
	return v.(T)
}

func (p *Pool[T]) Put(item T) {
	item.Reset()
	p.sp.Put(item)
}