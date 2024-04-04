package main

import (
	"sync"
	"unsafe"
)

/* GobMutex is a regular 'sync.Mutex', but it's designed to be compatible with 'encoding/gob'. */
type GobMutex [2]uint32

func (gm *GobMutex) Lock() {
	m := (*sync.Mutex)(unsafe.Pointer(gm))
	m.Lock()
}

func (gm *GobMutex) Unlock() {
	m := (*sync.Mutex)(unsafe.Pointer(gm))
	m.Unlock()
}
