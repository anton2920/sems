package main

import (
	"unsafe"
)

type EventQueue struct {
	platformEventQueue
}

type EventRequest int

const (
	EventRequestRead EventRequest = 1 << iota
	EventRequestWrite
)

type EventTrigger int

const (
	EventTriggerLevel EventTrigger = iota
	EventTriggerEdge
)

type Event interface{}

type ErrorEvent struct {
	Error error
}

type EndOfFileEvent struct {
	Handle   int32
	UserData unsafe.Pointer
}

type ReadEvent struct {
	Handle    int32
	Available int
	UserData  unsafe.Pointer
}

type WriteEvent struct {
	Handle    int32
	Available int
	UserData  unsafe.Pointer
}

type SignalEvent struct {
	Signal int32
}

func NewEventQueue() (*EventQueue, error) {
	q := new(EventQueue)
	if err := platformNewEventQueue(q); err != nil {
		return nil, err
	}
	return q, nil
}

func (q *EventQueue) AddSocket(l int32, request EventRequest, trigger EventTrigger, userData unsafe.Pointer) error {
	return platformQueueAddSocket(q, l, request, trigger, userData)
}

func (q *EventQueue) AddSignal(sig int32) error {
	return platformQueueAddSignal(q, sig)
}

func (q *EventQueue) Close() error {
	return platformQueueClose(q)
}

func (q *EventQueue) GetEvent() Event {
	return platformQueueGetEvent(q)
}

func (q *EventQueue) HasEvents() bool {
	return platformQueueHasEvents(q)
}
