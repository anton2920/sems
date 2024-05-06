package main

import "unsafe"

type Kevent_t struct {
	Ident  uintptr
	Filter int16
	Flags  uint16
	Fflags uint32
	Data   int
	Udata  unsafe.Pointer
	Ext    [4]uint
}

const (
	/* From <sys/event.h>. */
	EVFILT_READ   = -1
	EVFILT_WRITE  = -2
	EVFILT_VNODE  = -4
	EVFILT_SIGNAL = -6
	EVFILT_TIMER  = -7

	EV_ADD   = 0x0001
	EV_CLEAR = 0x0020

	EV_ERROR = 0x4000
	EV_EOF   = 0x8000

	NOTE_WRITE = 0x0002

	NOTE_SECONDS = 0x00000001
)
