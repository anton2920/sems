package main

import (
	"fmt"
	"unsafe"
)

type platformEventQueue struct {
	kq int32

	/* Events buffer */
	events [256]Kevent_t
	head   int
	tail   int
}

func platformNewEventQueue(q *EventQueue) error {
	kq, err := Kqueue()
	if err != nil {
		return fmt.Errorf("failed to open kernel queue: %w", err)
	}
	q.kq = kq
	return nil
}

func platformQueueAddSocket(q *EventQueue, l int32, request EventRequest, trigger EventTrigger, userData unsafe.Pointer) error {
	var flags uint16 = EV_ADD
	if trigger == EventTriggerEdge {
		flags |= EV_CLEAR
	}

	if (request & EventRequestRead) == EventRequestRead {
		event := Kevent_t{Ident: uintptr(l), Filter: EVFILT_READ, Flags: flags, Udata: userData}
		if _, err := Kevent(q.kq, unsafe.Slice(&event, 1), nil, nil); err != nil {
			return fmt.Errorf("failed to request socket read event: %w", err)
		}
	}

	if (request & EventRequestRead) == EventRequestRead {
		event := Kevent_t{Ident: uintptr(l), Filter: EVFILT_WRITE, Flags: flags, Udata: userData}
		if _, err := Kevent(q.kq, unsafe.Slice(&event, 1), nil, nil); err != nil {
			return fmt.Errorf("failed to request socket write event: %w", err)
		}
	}

	return nil
}

func platformQueueAddSignal(q *EventQueue, sig int32) error {
	event := Kevent_t{Ident: uintptr(sig), Filter: EVFILT_SIGNAL, Flags: EV_ADD}
	if _, err := Kevent(q.kq, unsafe.Slice(&event, 1), nil, nil); err != nil {
		return fmt.Errorf("failed to request signal event: %w", err)
	}

	return nil
}

func platformQueueClose(q *EventQueue) error {
	return Close(q.kq)
}

func platformQueueGetEvent(q *EventQueue) (Event, error) {
	var err error

	if q.head >= q.tail {
	retry:
		q.tail, err = Kevent(q.kq, nil, unsafe.Slice(&q.events[0], len(q.events)), nil)
		if err != nil {
			if err.(ErrorWithCode).Code == EINTR {
				goto retry
			}
			return EmptyEvent, err
		}
		q.head = 0
	}
	head := q.events[q.head]
	q.head++

	if (head.Flags & EV_ERROR) == EV_ERROR {
		return EmptyEvent, fmt.Errorf("requested event for %v failed with code %v", head.Ident, head.Data)
	}

	var event Event
	event.Identifier = int32(head.Ident)
	event.Available = head.Data
	event.UserData = head.Udata
	event.EndOfFile = (head.Flags & EV_EOF) == EV_EOF

	switch head.Filter {
	case EVFILT_READ:
		event.Type = EventRead
	case EVFILT_WRITE:
		event.Type = EventWrite
	case EVFILT_SIGNAL:
		event.Type = EventSignal
	}

	return event, nil
}

/* TODO(anton2920): think about returning the number of available events instead. */
/* BUG(anton2920): it doesn't work! */
func platformQueueHasEvents(q *EventQueue) bool {
	if q.head < q.tail {
		return true
	}

	n, err := Kevent(q.kq, nil, nil, nil)
	if err != nil {
		return false
	}

	return n > 0
}
