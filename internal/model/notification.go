package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationType string

const (
	MessageNotification NotificationType = "message"
	GroupNotification   NotificationType = "group"
	SystemNotification  NotificationType = "system"
)

type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type      NotificationType   `bson:"type" json:"type"`
	Title     string             `bson:"title" json:"title"`
	Content   string             `bson:"content" json:"content"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	SenderID  primitive.ObjectID `bson:"sender_id,omitempty" json:"sender_id,omitempty"`
	GroupID   primitive.ObjectID `bson:"group_id,omitempty" json:"group_id,omitempty"`
	IsRead    bool               `bson:"is_read" json:"is_read"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
