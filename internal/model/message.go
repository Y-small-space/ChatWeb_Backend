package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageType string

const (
	TextMessage  MessageType = "text"
	ImageMessage MessageType = "image"
	FileMessage  MessageType = "file"
)

type Message struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type       MessageType        `bson:"type" json:"type"`
	Content    string             `bson:"content" json:"content"`
	SenderID   primitive.ObjectID `bson:"sender_id" json:"sender_id"`
	ReceiverID primitive.ObjectID `bson:"receiver_id" json:"receiver_id"`
	GroupID    primitive.ObjectID `bson:"group_id,omitempty" json:"group_id,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
	Status     string             `bson:"status" json:"status"` // sent, delivered, read
	ReadBy     []ReadReceipt      `bson:"read_by" json:"read_by"`
}

type ReadReceipt struct {
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	ReadAt    time.Time          `bson:"read_at" json:"read_at"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
}
