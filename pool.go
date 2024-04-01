package main

import "unsafe"

type Pool struct {
	Items []unsafe.Pointer
	New   func() unsafe.Pointer
}

func NewPool(newF func() unsafe.Pointer) *Pool {
	ret := new(Pool)
	ret.Items = make([]unsafe.Pointer, 0, 1024)
	ret.New = newF
	return ret
}

func (p *Pool) Get() unsafe.Pointer {
	var item unsafe.Pointer

	if len(p.Items) > 0 {
		item = p.Items[len(p.Items)-1]
		p.Items = p.Items[:len(p.Items)-1]
	} else {
		item = p.New()
	}
	return item
}

func (p *Pool) Put(item unsafe.Pointer) {
	p.Items = append(p.Items, item)
}
