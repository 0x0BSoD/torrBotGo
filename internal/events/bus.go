// Package events provides an event bus system for torrBotGo.
// It enables publish-subscribe communication between components,
// allowing decoupled event handling for system notifications.
package events

import (
	"context"
	"sync"
)

type Handler func(Event)

type Bus struct {
	mu       sync.RWMutex
	handlers map[Type][]Handler
	ch       chan Event
}

func New(buffer int) *Bus {
	return &Bus{
		handlers: make(map[Type][]Handler),
		ch:       make(chan Event, buffer),
	}
}

func (b *Bus) Subscribe(t Type, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[t] = append(b.handlers[t], h)
}

func (b *Bus) Publish(ev Event) {
	b.ch <- ev
}

func (b *Bus) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-b.ch:
			b.dispatch(ev)
		}
	}
}

func (b *Bus) dispatch(ev Event) {
	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers[ev.Type]...)
	b.mu.RUnlock()

	for _, h := range handlers {
		go h(ev)
	}
}
