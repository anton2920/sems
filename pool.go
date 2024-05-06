package main

type Pool[T any] struct {
	Items   []*T
	NewItem func() (*T, error)
}

func NewPool[T any](newItem func() (*T, error)) Pool[T] {
	var p Pool[T]
	p.Items = make([]*T, 0, 1024)
	p.NewItem = newItem
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
	p.Items = append(p.Items, item)
}
