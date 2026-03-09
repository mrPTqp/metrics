package pool

import "testing"

type valueItem struct {
	value int
}


func (v valueItem) Reset() {}

type ptrItem struct {
	value int
}


func (p *ptrItem) Reset() {
	p.value = 0
}

func TestPool_GetReturnsZeroValue(t *testing.T) {
	p := NewPool[valueItem]()

	item := p.Get()
	if item.value != 0 {
		t.Errorf("new item value = %d, want %d", item.value, 0)
	}
}

func TestPool_PutResetsItem(t *testing.T) {
	p := NewPool[*ptrItem]()

	item := &ptrItem{value: 42}
	p.Put(item)

	if item.value != 0 {
		t.Errorf("item value after Put = %d, want %d", item.value, 0)
	}
}


