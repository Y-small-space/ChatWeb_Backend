package websocket

import (
	"encoding/json"
	"sync"

	"chatweb/pkg/event"
)

// Hub 管理所有活跃的WebSocket连接
type Hub struct {
	// 所有活跃的连接
	clients map[string]*Client
	// 用于保护clients map的互斥锁
	mu sync.RWMutex
	// 广播消息的通道
	broadcast chan []byte
	// 事件总线
	eventBus *event.EventBus
}

func NewHub(eventBus *event.EventBus) *Hub {
	hub := &Hub{
		clients:   make(map[string]*Client),
		broadcast: make(chan []byte),
		eventBus:  eventBus,
	}

	// 订阅相关事件
	hub.subscribeToEvents()
	return hub
}

func (h *Hub) subscribeToEvents() {
	h.eventBus.Subscribe(event.MessageRead, func(e event.Event) {
		if content, ok := e.Content.(event.MessageReadContent); ok {
			msg := struct {
				Type    string                   `json:"type"`
				Content event.MessageReadContent `json:"content"`
			}{
				Type:    "read",
				Content: content,
			}
			if messageBytes, err := json.Marshal(msg); err == nil {
				h.broadcast <- messageBytes
			}
		}
	})

	h.eventBus.Subscribe(event.GroupRead, func(e event.Event) {
		if content, ok := e.Content.(event.GroupReadContent); ok {
			msg := struct {
				Type    string                 `json:"type"`
				Content event.GroupReadContent `json:"content"`
			}{
				Type:    "group_read",
				Content: content,
			}
			if messageBytes, err := json.Marshal(msg); err == nil {
				h.broadcast <- messageBytes
			}
		}
	})

	// ... 订阅其他事件
}

func (h *Hub) Run() {
	for {
		select {
		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client.id)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	h.clients[client.id] = client
	h.mu.Unlock()
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	if _, ok := h.clients[client.id]; ok {
		delete(h.clients, client.id)
		close(client.send)
	}
	h.mu.Unlock()
}

// 添加定向消息发送方法
func (h *Hub) SendToUser(userID string, message []byte) {
	h.mu.RLock()
	if client, ok := h.clients[userID]; ok {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client.id)
		}
	}
	h.mu.RUnlock()
}

// 添加群发消息方法
func (h *Hub) BroadcastToUsers(userIDs []string, message []byte) {
	h.mu.RLock()
	for _, userID := range userIDs {
		if client, ok := h.clients[userID]; ok {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client.id)
			}
		}
	}
	h.mu.RUnlock()
}
