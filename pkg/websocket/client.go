package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"chatweb/internal/service"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512

	MessageTypeChat         = "chat"
	MessageTypeTyping       = "typing"
	MessageTypeNotification = "notification"
	MessageTypeOnline       = "online"
	MessageTypeRead         = "read"
	MessageTypeGroupRead    = "group_read"
)

type Client struct {
	hub           *Hub
	conn          *websocket.Conn
	send          chan []byte
	id            string // 用户ID
	onlineService *service.OnlineService
}

type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

type OnlineStatusMessage struct {
	UserID   string `json:"user_id"`
	IsOnline bool   `json:"is_online"`
}

type ReadStatusMessage struct {
	MessageID string `json:"message_id"`
	UserID    string `json:"user_id"`
	ReadAt    string `json:"read_at"`
}

type GroupReadStatusMessage struct {
	MessageID  string   `json:"message_id"`
	GroupID    string   `json:"group_id"`
	ReadByUser string   `json:"read_by_user"`
	ReadAt     string   `json:"read_at"`
	ReadCount  int      `json:"read_count"`
	ReadBy     []string `json:"read_by"`
}

// Upgrader 配置
type Upgrader = websocket.Upgrader

var defaultUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 在生产环境中应该更严格
	},
}

func NewClient(hub *Hub, conn *websocket.Conn, userID string, onlineService *service.OnlineService) *Client {
	return &Client{
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		id:            userID,
		onlineService: onlineService,
	}
}

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

		// 处理消息
		c.handleMessage(msg)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.onlineService.SetUserOffline(context.Background(), c.id); err != nil {
			log.Printf("Failed to set user offline: %v", err)
		}
		c.conn.Close()
	}()

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

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg Message) {
	switch msg.Type {
	case MessageTypeChat:
		c.handleChatMessage(msg)
	case MessageTypeTyping:
		c.handleTypingStatus(msg)
	case MessageTypeNotification:
		c.handleNotification(msg)
	}
}

func (c *Client) handleChatMessage(msg Message) {
	// 将消息广播给所有连接的客户端
	messageBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error marshaling message: %v", err)
		return
	}
	c.hub.broadcast <- messageBytes
}

func (c *Client) handleTypingStatus(msg Message) {
	// 处理用户正在输入的状态
	messageBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error marshaling typing status: %v", err)
		return
	}
	c.hub.broadcast <- messageBytes
}

func (c *Client) handleNotification(msg Message) {
	messageBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error marshaling notification: %v", err)
		return
	}
	// 发送给指定用户
	if notification, ok := msg.Content.(map[string]interface{}); ok {
		if userID, ok := notification["user_id"].(string); ok {
			c.hub.SendToUser(userID, messageBytes)
			return
		}
	}
	// 如果没有指定用户，则广播
	c.hub.broadcast <- messageBytes
}

func (c *Client) SendMessage(msg interface{}) error {
	messageBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.send <- messageBytes:
		return nil
	default:
		return errors.New("send channel is full")
	}
}
