package transport

import (
	"context"

	"github.com/narcilee7/dew/pkg/event"
)

// EventStreamAdapter converts between core.EventStream and Go channels.
type EventStreamAdapter struct {
	ch     chan event.Event
	closed bool
}

// NewEventStreamAdapter creates a new adapter with a buffered channel.
func NewEventStreamAdapter(buf int) (*EventStreamAdapter, chan event.Event) {
	ch := make(chan event.Event, buf)
	return &EventStreamAdapter{ch: ch}, ch
}

// Recv receives the next event from the stream.
func (a *EventStreamAdapter) Recv() (event.Event, error) {
	ev, ok := <-a.ch
	if !ok {
		return nil, context.Canceled
	}
	return ev, nil
}

// Close closes the stream.
func (a *EventStreamAdapter) Close() error {
	if !a.closed {
		close(a.ch)
		a.closed = true
	}
	return nil
}
