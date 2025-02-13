package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Friendship 定义用户之间的好友关系
type Friendship struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`      // 好友关系的唯一标识符
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`       // 用户的 ID
	FriendID  primitive.ObjectID `bson:"friend_id" json:"friend_id"`   // 好友的 ID
	CreatedAt time.Time          `bson:"created_at" json:"created_at"` // 关系创建时间
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"` // 关系更新时间
}
