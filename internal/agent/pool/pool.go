package pool

type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	items []T
}

func NewPool[T Resetter]() *Pool[T] {
	return &Pool[T]{
		items: make([]T, 0),
	}
}

func (p *Pool[T]) Get() T {
	if len(p.items) == 0 {
		var r T
		return r
	}
	last := p.items[len(p.items)-1]
	p.items = p.items[:len(p.items)-1]

	return last
}

func (p *Pool[T]) Put(item T) {
	item.Reset()
	p.items = append(p.items, item)
}
