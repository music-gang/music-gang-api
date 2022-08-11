package event_test

import (
	"context"
	"testing"

	"github.com/music-gang/music-gang-api/event"
)

const (
	EventTest1Type event.EventType = "eventTest1"
	EventTest2Type event.EventType = "eventTest2"
)

func TestEventService(t *testing.T) {

	ctx := context.Background()

	s := event.NewEventService()

	sub1 := s.Subscribe(ctx, EventTest1Type)
	sub2 := s.Subscribe(ctx, EventTest2Type)

	s.PublishEvent(ctx, event.Event{
		Type:    EventTest1Type,
		Message: "test1",
		Payload: true,
	})

	s.PublishEvent(ctx, event.Event{
		Type:    EventTest2Type,
		Message: "test2",
		Payload: false,
	})

	select {
	case e := <-sub1.C():
		if e.Type != EventTest1Type {
			t.Errorf("expected event type %s, got %s", EventTest1Type, e.Type)
		} else if e.Message != "test1" {
			t.Errorf("expected event message %s, got %s", "test1", e.Message)
		} else if e.Payload.(bool) != true {
			t.Errorf("expected event payload %v, got %v", true, e.Payload)
		}
	default:
		t.Error("expected event")
	}

	select {
	case e := <-sub2.C():
		if e.Type != EventTest2Type {
			t.Errorf("expected event type %s, got %s", EventTest2Type, e.Type)
		} else if e.Message != "test2" {
			t.Errorf("expected event message %s, got %s", "test2", e.Message)
		} else if e.Payload.(bool) != false {
			t.Errorf("expected event payload %v, got %v", false, e.Payload)
		}
	default:
		t.Error("expected event")
	}

	select {
	case e := <-sub1.C():
		t.Errorf("unexpected event %v", e)
	case e := <-sub2.C():
		t.Errorf("unexpected event %v", e)
	default:
	}

	sub1.Close()
	sub2.Close()

	// now, try to publish an event to a context closed subscription

	ctx, cancel := context.WithCancel(ctx)

	sub3 := s.Subscribe(ctx, EventTest1Type)

	cancel()

	s.PublishEvent(ctx, event.Event{
		Type:    EventTest1Type,
		Message: "test1-context-closed",
		Payload: true,
	})

	if e, ok := <-sub3.C(); ok {
		t.Errorf("unexpected event %v", e)
	}
}
