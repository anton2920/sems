package main

import (
	"sync"
	"unsafe"

	"github.com/anton2920/gofa/prof"
)

/* GobMutex is a regular 'sync.Mutex', but it's designed to be compatible with 'encoding/gob'. */
type GobMutex [2]uint32

func (gm *GobMutex) Lock() {
	defer prof.End(prof.Begin(""))

	m := (*sync.Mutex)(unsafe.Pointer(gm))
	m.Lock()
}

func (gm *GobMutex) Unlock() {
	defer prof.End(prof.Begin(""))

	m := (*sync.Mutex)(unsafe.Pointer(gm))
	m.Unlock()
}
