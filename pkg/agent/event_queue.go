package agent

import (
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type Operation int

const (
	Pull Operation = iota
	Remove
)

type Event struct {
	Image     string
	Auth      []*runtime.AuthConfig
	Operation Operation
}

type EventQueue chan *Event

func NewEventQueue() EventQueue {
	return make(chan *Event)
}

func (eq EventQueue) Close() {
	close(eq)
}
