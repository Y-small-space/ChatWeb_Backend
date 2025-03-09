package websocketM

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"chatweb/internal/model"
	"chatweb/internal/service"

	"chatweb/internal/repository"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// WebSocket 相关超时时间和消息大小限制
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10 // 心跳检测时间间隔
	maxMessageSize = 512                 // 最大消息大小

	// 定义 WebSocket 消息类型
	MessageTypeChat         = "chat"
	MessageTypeTyping       = "typing"
	MessageTypeNotification = "notification"
	MessageTypeOnline       = "online"
	MessageTypeRead         = "read"
	MessageTypeGroupRead    = "group_read"
)

// Client 代表一个 WebSocket 连接的客户端
// 负责管理连接、读取和发送消息

type Client struct {
	hub            *Hub                    // WebSocket 集线器
	conn           *websocket.Conn         // WebSocket 连接实例
	send           chan []byte             // 发送消息的通道
	id             string                  // 客户端的用户 ID
	onlineService  *service.OnlineService  // 在线状态服务
	messageService *service.MessageService // 消息服务
}
type MessageType string

// Message 结构体用于解析 WebSocket 消息
type Message struct {
	Type       MessageType        `bson:"type" json:"type"`                             // 消息类型（文本、图片、文件）
	Content    string             `bson:"content" json:"content"`                       // 消息内容
	SenderID   primitive.ObjectID `bson:"sender_id" json:"sender_id"`                   // 发送者的用户 ID
	ReceiverID primitive.ObjectID `bson:"receiver_id" json:"receiver_id"`               // 接收者的用户 ID
	GroupID    primitive.ObjectID `bson:"group_id,omitempty" json:"group_id,omitempty"` // 群组 ID（如果是群消息）
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`                 // 消息发送时间
	Sender     string             `bson:"sender" json:"sender"`                         // 发送者 name
	Receiver   string             `bson:"receiver" json:"receiver"`                     // 接收者 name
}

// OnlineStatusMessage 结构体用于用户在线状态的消息

type OnlineStatusMessage struct {
	UserID   string `json:"user_id"`
	IsOnline bool   `json:"is_online"`
}

// ReadStatusMessage 结构体用于已读状态消息

type ReadStatusMessage struct {
	MessageID string `json:"message_id"`
	UserID    string `json:"user_id"`
	ReadAt    string `json:"read_at"`
}

// GroupReadStatusMessage 结构体用于群组已读状态消息

type GroupReadStatusMessage struct {
	MessageID  string   `json:"message_id"`
	GroupID    string   `json:"group_id"`
	ReadByUser string   `json:"read_by_user"`
	ReadAt     string   `json:"read_at"`
	ReadCount  int      `json:"read_count"`
	ReadBy     []string `json:"read_by"`
}

// WebSocket 连接升级配置
var defaultUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境应实现更严格的安全检查
	},
}

// NewClient 创建新的 WebSocket 客户端
func NewClient(hub *Hub, conn *websocket.Conn, userID string, onlineService *service.OnlineService) *Client {
	return &Client{
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		id:            userID,
		onlineService: onlineService,
	}
}

// ReadPump 监听 WebSocket 连接的读取操作
// 处理消息并将其转发到相应的处理器
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		if err := c.onlineService.SetUserOffline(context.Background(), c.id); err != nil {
			log.Printf("Failed to set user offline: %v", err)
		}
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("error unmarshaling message: %v", err)
			continue
		}
		// 处理不同类型的消息
		c.handleMessage(msg)
	}
}

// WritePump 负责向客户端发送消息
// 包括定期发送心跳包以保持连接活跃
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	// defer func() {
	// 	ticker.Stop()
	// 	if err := c.onlineService.SetUserOffline(context.Background(), c.id); err != nil {
	// 		log.Printf("Failed to set user offline: %v", err)
	// 	}
	// 	c.conn.Close()
	// }()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			w.Close()

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 处理不同类型的 WebSocket 消息
func (c *Client) handleMessage(msg Message) {
	c.handleChatMessage(msg)
}

// 处理聊天消息，广播给所有客户端
func (c *Client) handleChatMessage(msg Message) {
	messageBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error marshaling message: %v", err)
		return
	}
	log.Printf(msg.Content, msg.SenderID)
	log.Printf("hanleChatMessage: %v", msg)

	// 构建消息对象
	message := &model.Message{
		Type:       model.MessageType(msg.Type),
		Content:    msg.Content,
		SenderID:   msg.SenderID,
		ReceiverID: msg.ReceiverID,
		Sender:     msg.Sender,
		Receiver:   msg.Receiver,
		Status:     "sent",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if msg.GroupID.String() != "" {
		message.GroupID = msg.GroupID
	}

	// 保存消息到数据库（异步执行，不阻塞 WebSocket）
	// go func() {
	// 	// 假设你有一个 MessageService 实例作为数据库操作服务
	// 	if err := c.messageService.CreateMessage(context.Background(), message); err != nil {
	// 		log.Printf("Failed to insert message into DB: %v", err)
	// 	}
	// }()

	repository.NewMessageRepository().Create(context.Background(), message)

	// 如果有接收者ID，发送给目标客户端
	if msg.ReceiverID.String() != "" {
		log.Println(c.hub.clients)
		log.Println(msg.ReceiverID)
		log.Println(msg.ReceiverID.Hex())
		id := msg.ReceiverID.Hex()
		recipientClient, ok := c.hub.clients[id]
		if ok {
			recipientClient.send <- messageBytes
		} else {
			log.Printf("Recipient %s not found", msg.ReceiverID.String())
		}
	} else if msg.GroupID.String() != "" {
		// 如果有群组ID，发送给群组内所有客户端
		for _, client := range c.hub.clients {
			if client.id == c.id {
				continue
			}
			if client.id == c.id {
				continue
			}
			client.send <- messageBytes
		}
	} else {
		// 否则广播给所有客户端
		c.hub.broadcast <- messageBytes
	}
}

// 处理用户输入状态
// func (c *Client) handleTypingStatus(msg Message) {
// 	messageBytes, err := json.Marshal(msg)
// 	if err != nil {
// 		log.Printf("error marshaling typing status: %v", err)
// 		return
// 	}
// 	c.hub.broadcast <- messageBytes
// }

// 处理通知消息
// func (c *Client) handleNotification(msg Message) {
// 	messageBytes, err := json.Marshal(msg)
// 	if err != nil {
// 		log.Printf("error marshaling notification: %v", err)
// 		return
// 	}
// 	c.hub.broadcast <- messageBytes
// }
