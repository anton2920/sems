package main

import (
	"runtime"
	"unsafe"
)

type EventQueue struct {
	platformEventQueue

	Pinner    runtime.Pinner
	LastPause int64
}

type EventRequest int

const (
	EventRequestRead EventRequest = (1 << iota)
	EventRequestWrite
)

type EventTrigger int

const (
	EventTriggerLevel EventTrigger = iota
	EventTriggerEdge
)

type EventType int32

const (
	EventNone EventType = iota
	EventRead
	EventWrite
	EventSignal
)

type Event struct {
	Type       EventType
	Identifier int32
	Available  int
	UserData   unsafe.Pointer

	/* TODO(anton2920): I don't like this!!! */
	EndOfFile bool
}

var EmptyEvent Event

func NewEventQueue() (*EventQueue, error) {
	q := new(EventQueue)
	if err := platformNewEventQueue(q); err != nil {
		return nil, err
	}
	return q, nil
}

func (q *EventQueue) AddHTTPClient(ctx *HTTPContext, request EventRequest, trigger EventTrigger) error {
	q.Pinner.Pin(ctx)
	return platformQueueAddSocket(q, ctx.Connection, request, trigger, unsafe.Pointer(uintptr(unsafe.Pointer(ctx))|uintptr(ctx.Check)))
}

func (q *EventQueue) AddSocket(sock int32, request EventRequest, trigger EventTrigger, userData unsafe.Pointer) error {
	return platformQueueAddSocket(q, sock, request, trigger, userData)
}

func (q *EventQueue) AddSignal(sig int32) error {
	return platformQueueAddSignal(q, sig)
}

func (q *EventQueue) Close() error {
	q.Pinner.Unpin()
	return platformQueueClose(q)
}

func (q *EventQueue) GetEvent() (Event, error) {
	return platformQueueGetEvent(q)
}

func (q *EventQueue) HasEvents() bool {
	return platformQueueHasEvents(q)
}

func (q *EventQueue) Pause(FPS int) {
	now := platformQueueGetTime()
	durationBetweenPauses := now - q.LastPause
	targetRate := int64(1000.0/float32(FPS)) * 1_000_000

	duration := targetRate - durationBetweenPauses
	if duration > 0 {
		platformQueuePause(q, duration)
		now = platformQueueGetTime()
	}
	q.LastPause = now
}
