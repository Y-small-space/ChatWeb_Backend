package event

import (
	"sync"
)

// EventType 定义事件的类型
type EventType string

// 定义几种常见的事件类型
const (
	MessageSent  EventType = "message_sent" // 消息已发送
	MessageRead  EventType = "message_read" // 消息已读
	GroupRead    EventType = "group_read"   // 群组消息已读
	UserOnline   EventType = "user_online"  // 用户上线
	UserOffline  EventType = "user_offline" // 用户下线
	Notification EventType = "notification" // 通知
)

// Event 表示一个事件的结构
type Event struct {
	Type    EventType   `json:"type"`    // 事件类型
	Content interface{} `json:"content"` // 事件内容，可以是任意类型
}

// 预定义的事件内容结构

// MessageReadContent 表示消息已读事件的内容
type MessageReadContent struct {
	MessageID string `json:"message_id"` // 消息ID
	UserID    string `json:"user_id"`    // 用户ID
	ReadAt    string `json:"read_at"`    // 阅读时间
	IsGroup   bool   `json:"is_group"`   // 是否是群组消息
}

// GroupReadContent 表示群组消息已读事件的内容
type GroupReadContent struct {
	MessageID  string   `json:"message_id"`   // 消息ID
	GroupID    string   `json:"group_id"`     // 群组ID
	ReadByUser string   `json:"read_by_user"` // 已读的用户ID
	ReadAt     string   `json:"read_at"`      // 阅读时间
	ReadCount  int      `json:"read_count"`   // 阅读人数
	ReadBy     []string `json:"read_by"`      // 已阅读的用户列表
}

// UserStatusContent 表示用户状态变化事件的内容
type UserStatusContent struct {
	UserID   string `json:"user_id"`   // 用户ID
	IsOnline bool   `json:"is_online"` // 是否在线
}

// Handler 定义了事件处理函数的类型
type Handler func(event Event)

// EventBus 事件总线结构
type EventBus struct {
	handlers map[EventType][]Handler // 存储事件类型与其处理函数的映射
	mu       sync.RWMutex            // 用于并发控制的读写锁
}

// NewEventBus 创建一个新的 EventBus 实例
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]Handler), // 初始化事件处理函数映射
	}
}

// Subscribe 订阅一个特定类型的事件
func (b *EventBus) Subscribe(eventType EventType, handler Handler) {
	b.mu.Lock()         // 加锁以保证线程安全
	defer b.mu.Unlock() // 解锁

	// 将事件处理函数添加到指定事件类型的处理列表中
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

// Publish 发布一个事件
func (b *EventBus) Publish(event Event) {
	b.mu.RLock()         // 读锁，确保并发安全
	defer b.mu.RUnlock() // 解锁

	// 查找并触发该事件类型的所有处理函数
	if handlers, ok := b.handlers[event.Type]; ok {
		// 并发执行所有处理函数
		for _, handler := range handlers {
			go handler(event)
		}
	}
}
