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

type KqueueCb func(Kevent_t) error

const (
	/* From <sys/event.h>. */
	EVFILT_READ  = -1
	EVFILT_WRITE = -2
	EVFILT_VNODE = -4
	EVFILT_TIMER = -7

	EV_ADD   = 0x0001
	EV_CLEAR = 0x0020

	EV_EOF = 0x8000

	NOTE_WRITE = 0x0002

	NOTE_SECONDS = 0x00000001
)

func KqueueMonitor(eventlist []Kevent_t, cb KqueueCb) error {
	kq, err := Kqueue()
	if err != nil {
		return err
	}

	if _, err := Kevent(kq, eventlist, nil, nil); err != nil {
		return err
	}

	var event Kevent_t
	for {
		if _, err := Kevent(kq, nil, unsafe.Slice(&event, 1), nil); err != nil {
			code := err.(E).Code
			if code != EINTR {
				return err
			}
			continue
		} else if err := cb(event); err != nil {
			return err
		}

		/* NOTE(anton2920): sleep to prevent runaway events. */
		SleepFull(Timespec{Nsec: 200000000})
	}
}
