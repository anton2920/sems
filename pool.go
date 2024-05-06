package main

type Pool[T any] struct {
	Items   []*T
	NewItem func() (*T, error)
	Reset   func(*T)
}

func NewPool[T any](newItem func() (*T, error), reset func(*T)) Pool[T] {
	var p Pool[T]
	p.Items = make([]*T, 0, 1024)
	p.NewItem = newItem
	p.Reset = reset
	return p
}

func (p *Pool[T]) Get() (*T, error) {
	if len(p.Items) == 0 {
		return p.NewItem()
	}

	item := p.Items[len(p.Items)-1]
	p.Items = p.Items[:len(p.Items)-1]
	return item, nil
}

func (p *Pool[T]) Put(item *T) {
	p.Reset(item)
	p.Items = append(p.Items, item)
}
