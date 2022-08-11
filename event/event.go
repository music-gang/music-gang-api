package event

import (
	"context"
	"sync"
)

// EventType is the type of the event.
type EventType string

const (
	EngineShouldResumeEvent EventType = "engineShouldResume"
	EngineShouldPauseEvent  EventType = "engineShouldPause"
)

// EventBufferSize is the size of the event buffer of the channel for each subscribers.
const EventBufferSize = 16

// EventService is a service for publishing and subscribing to events.
type EventService struct {
	mux  sync.Mutex
	subs map[EventType]map[*Subscription]struct{}
}

// NewEventService creates a new event service.
func NewEventService() *EventService {
	return &EventService{
		subs: make(map[EventType]map[*Subscription]struct{}),
	}
}

// PublishEvent publishes the given event to all subscribers.
func (s *EventService) PublishEvent(ctx context.Context, event Event) {
	s.mux.Lock()
	defer s.mux.Unlock()

	subs, ok := s.subs[event.Type]
	if !ok {
		return
	}

	for sub := range subs {

		select {
		case <-sub.ctx.Done():
			s.unsubscribe(sub)
			continue
		default:
		}

		select {
		case sub.c <- event:
		default:
			s.unsubscribe(sub)

		}
	}
}

// Subscribe subscribes to the given event.
// Context is used to automatically unsubscribe the subscription when it is canceled.
func (s *EventService) Subscribe(ctx context.Context, event EventType) (sub *Subscription) {

	sub = &Subscription{
		ctx:          ctx,
		unsubService: s,
		eventType:    event,
		c:            make(chan Event, EventBufferSize),
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	if _, ok := s.subs[event]; !ok {
		s.subs[event] = make(map[*Subscription]struct{})
	}

	s.subs[event][sub] = struct{}{}

	return sub
}

// Unsubscribe unsubscribes the given subscription.
// It is safe to call this method concurrently.
func (s *EventService) Unsubscribe(sub *Subscription) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.unsubscribe(sub)
}

// unsubscribe is the actual implementation of Unsubscribe.
// It is should be called while holding the lock.
// It is not safe to call this method concurrently.
func (s *EventService) unsubscribe(sub *Subscription) {

	sub.onceClose.Do(func() {
		close(sub.c)
	})

	subs, ok := s.subs[sub.eventType]
	if !ok {
		return
	}

	delete(subs, sub)

	if len(subs) == 0 {
		delete(s.subs, sub.eventType)
	}
}

// Subscription is a subscription to an event service.
type Subscription struct {
	ctx          context.Context
	eventType    EventType
	unsubService *EventService
	c            chan Event
	onceClose    sync.Once
}

// Ctx returns the context of the subscription.
func (s *Subscription) Ctx() context.Context {
	return s.ctx
}

// Close unsubscribes the subscription.
func (s *Subscription) Close() {
	s.unsubService.Unsubscribe(s)
}

// C returns the channel of the subscription.
func (s *Subscription) C() <-chan Event {
	return s.c
}

// Event is the event that is published to the subscription.
type Event struct {
	Type    EventType `json:"type"`
	Message string    `json:"message"`
	Payload any       `json:"payload"`
}
