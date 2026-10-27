package service

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cuongbuilh/ai-business-os/backend/internal/domain"
)

type EventBroker struct {
	mu          sync.RWMutex
	subscribers map[chan string]bool
}

func NewEventBroker() *EventBroker {
	return &EventBroker{
		subscribers: make(map[chan string]bool),
	}
}

func (b *EventBroker) Subscribe() chan string {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan string, 32)
	b.subscribers[ch] = true
	return ch
}

func (b *EventBroker) Unsubscribe(ch chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subscribers[ch] {
		delete(b.subscribers, ch)
		close(ch)
	}
}

func (b *EventBroker) Publish(eventType, action string, payload interface{}) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	msg := domain.EventMessage{
		Type:      eventType,
		Action:    action,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	formatted := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(data))

	for ch := range b.subscribers {
		select {
		case ch <- formatted:
		default:
			// Non-blocking drop if slow consumer
		}
	}
}
