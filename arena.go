package main

type Arena struct {
}

func (a *Arena) NewSlice(n int) []byte {
	return make([]byte, n)
}
