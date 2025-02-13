package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Group 定义群组的数据结构
type Group struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`  // 群组的唯一标识符
	Name        string               `bson:"name" json:"name"`  // 群组名称
	Description string               `bson:"description" json:"description"`  // 群组描述
	CreatorID   primitive.ObjectID   `bson:"creator_id" json:"creator_id"`  // 群组创建者的 ID
	Members     []primitive.ObjectID `bson:"members" json:"members"`  // 群组成员列表
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`  // 群组创建时间
	UpdatedAt   time.Time            `bson:"updated_at" json:"updated_at"`  // 群组更新时间
}

// GroupMember 定义群组成员的数据结构
type GroupMember struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`  // 群组成员的唯一标识符
	GroupID   primitive.ObjectID `bson:"group_id" json:"group_id"`  // 所在群组的 ID
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`  // 成员的用户 ID
	Role      string             `bson:"role" json:"role"`  // 成员角色（admin 或 member）
	JoinedAt  time.Time          `bson:"joined_at" json:"joined_at"`  // 成员加入群组的时间
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`  // 成员信息更新时间
}