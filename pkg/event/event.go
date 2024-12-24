package event

import (
	"sync"
)

type EventType string

const (
	MessageSent  EventType = "message_sent"
	MessageRead  EventType = "message_read"
	GroupRead    EventType = "group_read"
	UserOnline   EventType = "user_online"
	UserOffline  EventType = "user_offline"
	Notification EventType = "notification"
)

type Event struct {
	Type    EventType   `json:"type"`
	Content interface{} `json:"content"`
}

// 预定义的事件内容结构
type MessageReadContent struct {
	MessageID string `json:"message_id"`
	UserID    string `json:"user_id"`
	ReadAt    string `json:"read_at"`
	IsGroup   bool   `json:"is_group"`
}

type GroupReadContent struct {
	MessageID  string   `json:"message_id"`
	GroupID    string   `json:"group_id"`
	ReadByUser string   `json:"read_by_user"`
	ReadAt     string   `json:"read_at"`
	ReadCount  int      `json:"read_count"`
	ReadBy     []string `json:"read_by"`
}

type UserStatusContent struct {
	UserID   string `json:"user_id"`
	IsOnline bool   `json:"is_online"`
}

type Handler func(event Event)

type EventBus struct {
	handlers map[EventType][]Handler
	mu       sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]Handler),
	}
}

func (b *EventBus) Subscribe(eventType EventType, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *EventBus) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if handlers, ok := b.handlers[event.Type]; ok {
		for _, handler := range handlers {
			go handler(event)
		}
	}
}
