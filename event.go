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

func (q *EventQueue) AddSocket(l int32, request EventRequest, trigger EventTrigger, userData unsafe.Pointer) error {
	return platformQueueAddSocket(q, l, request, trigger, userData)
}

func (q *EventQueue) AddSignal(sig int32) error {
	return platformQueueAddSignal(q, sig)
}

func (q *EventQueue) Close() error {
	return platformQueueClose(q)
}

func (q *EventQueue) GetEvent() (Event, error) {
	return platformQueueGetEvent(q)
}

func (q *EventQueue) HasEvents() bool {
	return platformQueueHasEvents(q)
}
