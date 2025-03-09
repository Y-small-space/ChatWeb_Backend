package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageType string

const (
	TextMessage  MessageType = "text"  // 文本消息
	ImageMessage MessageType = "image" // 图片消息
	FileMessage  MessageType = "file"  // 文件消息
)

// Message 定义消息的数据结构
type Message struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`                      // 消息的唯一标识符
	Type       MessageType        `bson:"type" json:"type"`                             // 消息类型（文本、图片、文件）
	Content    string             `bson:"content" json:"content"`                       // 消息内容
	SenderID   primitive.ObjectID `bson:"sender_id" json:"sender_id"`                   // 发送者的用户 ID
	ReceiverID primitive.ObjectID `bson:"receiver_id" json:"receiver_id"`               // 接收者的用户 ID
	Sender     string             `bson:"sender" json:"sender"`                         // 发送者 name
	Receiver   string             `bson:"receiver" json:"receiverer"`                   // 接收者 name
	GroupID    primitive.ObjectID `bson:"group_id,omitempty" json:"group_id,omitempty"` // 群组 ID（如果是群消息）
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`                 // 消息发送时间
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`                 // 消息更新时间
	Status     string             `bson:"status" json:"status"`                         // 消息状态（sent, delivered, read）
	ReadBy     []ReadReceipt      `bson:"read_by" json:"read_by"`                       // 读取消息的用户列表
	// Transfrom  string
}

// ReadReceipt 定义消息已读回执
type ReadReceipt struct {
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`     // 已读用户的 ID
	ReadAt    time.Time          `bson:"read_at" json:"read_at"`     // 阅读时间
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"` // 消息发送的时间戳
}
