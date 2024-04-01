package main

import "unsafe"

/* See <sys/_iovec.h>. */
type Iovec struct {
	Base unsafe.Pointer
	Len  uint64
}

func IovecForByteSlice(buf []byte) Iovec {
	return Iovec{Base: unsafe.Pointer(unsafe.SliceData(buf)), Len: uint64(len(buf))}
}

func IovecForString(s string) Iovec {
	return Iovec{Base: unsafe.Pointer(unsafe.StringData(s)), Len: uint64(len(s))}
}
