package websocketM

import (
	"encoding/json"
	"sync"

	"chatweb/pkg/event"
)

// Hub 管理所有活跃的 WebSocket 连接
type Hub struct {
	// clients 是一个 map，用于存储所有当前连接的 WebSocket 客户端
	clients map[string]*Client
	// mu 是一个互斥锁，用于保证对 clients map 的并发读写安全
	mu sync.RWMutex
	// broadcast 是一个通道，用于广播消息到所有连接的客户端
	broadcast chan []byte
	// eventBus 是事件总线，用于订阅和发布不同的事件
	eventBus *event.EventBus
}

// NewHub 创建一个新的 Hub 实例，初始化相关字段
func NewHub(eventBus *event.EventBus) *Hub {
	// 初始化 Hub 实例
	hub := &Hub{
		clients:   make(map[string]*Client),
		broadcast: make(chan []byte), // 广播通道
		eventBus:  eventBus,          // 事件总线
	}

	// 订阅相关事件
	hub.subscribeToEvents()
	return hub
}

// subscribeToEvents 订阅事件总线中的相关事件并处理消息
func (h *Hub) subscribeToEvents() {
	// 订阅消息已读事件
	h.eventBus.Subscribe(event.MessageRead, func(e event.Event) {
		if content, ok := e.Content.(event.MessageReadContent); ok {
			// 构造消息格式
			msg := struct {
				Type    string                   `json:"type"`
				Content event.MessageReadContent `json:"content"`
			}{
				Type:    "read",
				Content: content,
			}
			// 将消息序列化并广播
			if messageBytes, err := json.Marshal(msg); err == nil {
				h.broadcast <- messageBytes
			}
		}
	})

	// 订阅群聊已读事件
	h.eventBus.Subscribe(event.GroupRead, func(e event.Event) {
		if content, ok := e.Content.(event.GroupReadContent); ok {
			// 构造消息格式
			msg := struct {
				Type    string                 `json:"type"`
				Content event.GroupReadContent `json:"content"`
			}{
				Type:    "group_read",
				Content: content,
			}
			// 将消息序列化并广播
			if messageBytes, err := json.Marshal(msg); err == nil {
				h.broadcast <- messageBytes
			}
		}
	})

	// 可以在此继续订阅其他事件
}

// Run 启动 Hub，监听广播通道并将消息发送到所有连接的客户端
func (h *Hub) Run() {
	for {
		select {
		case message := <-h.broadcast: // 收到广播消息
			// 使用读锁保护 clients map
			h.mu.RLock()
			// 遍历所有客户端并将消息发送过去
			for _, client := range h.clients {
				select {
				case client.send <- message: // 如果客户端连接正常，发送消息
				default:
					// 如果客户端连接异常，关闭连接并从 clients 中删除
					close(client.send)
					delete(h.clients, client.id)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register 将客户端注册到 Hub 中
func (h *Hub) Register(client *Client) {
	h.mu.Lock() // 获取写锁
	h.clients[client.id] = client
	h.mu.Unlock() // 释放写锁
}

// Unregister 从 Hub 中注销客户端
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock() // 获取写锁
	if _, ok := h.clients[client.id]; ok {
		// 删除客户端并关闭发送通道
		delete(h.clients, client.id)
		close(client.send)
	}
	h.mu.Unlock() // 释放写锁
}

// SendToUser 发送定向消息给指定用户
func (h *Hub) SendToUser(userID string, message []byte) {
	h.mu.RLock() // 获取读锁
	if client, ok := h.clients[userID]; ok {
		// 如果该用户存在，发送消息
		select {
		case client.send <- message:
		default:
			// 如果客户端连接异常，关闭连接并从 clients 中删除
			close(client.send)
			delete(h.clients, client.id)
		}
	}
	h.mu.RUnlock() // 释放读锁
}

// BroadcastToUsers 向指定用户列表广播消息
func (h *Hub) BroadcastToUsers(userIDs []string, message []byte) {
	h.mu.RLock() // 获取读锁
	for _, userID := range userIDs {
		// 遍历用户列表，向每个用户发送消息
		if client, ok := h.clients[userID]; ok {
			select {
			case client.send <- message:
			default:
				// 如果客户端连接异常，关闭连接并从 clients 中删除
				close(client.send)
				delete(h.clients, client.id)
			}
		}
	}
	h.mu.RUnlock() // 释放读锁
}
