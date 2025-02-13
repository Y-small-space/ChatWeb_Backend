package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationType 定义通知类型
type NotificationType string

const (
	MessageNotification NotificationType = "message" // 消息通知
	GroupNotification   NotificationType = "group"   // 群组通知
	SystemNotification  NotificationType = "system"  // 系统通知
)

// Notification 定义通知的数据结构
type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`                        // 通知的唯一标识符
	Type      NotificationType   `bson:"type" json:"type"`                               // 通知类型
	Title     string             `bson:"title" json:"title"`                             // 通知标题
	Content   string             `bson:"content" json:"content"`                         // 通知内容
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`                         // 用户 ID
	SenderID  primitive.ObjectID `bson:"sender_id,omitempty" json:"sender_id,omitempty"` // 发送者 ID
	GroupID   primitive.ObjectID `bson:"group_id,omitempty" json:"group_id,omitempty"`   // 群组 ID
	IsRead    bool               `bson:"is_read" json:"is_read"`                         // 是否已读
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`                   // 通知创建时间
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`                   // 通知更新时间
}
