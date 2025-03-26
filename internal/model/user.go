package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User 定义用户的数据结构
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`      // 用户的唯一标识符
	Username  string             `bson:"username" json:"username"`     // 用户名
	Email     string             `bson:"email" json:"email"`           // 用户邮箱
	Password  string             `bson:"password" json:"-"`            // 用户密码（不序列化）
	Phone     string             `bson:"phone" json:"phone"`           // 用户手机号码
	Avatar    string             `bson:"avatar" json:"avatar"`         // 头像url
	CreatedAt time.Time          `bson:"created_at" json:"created_at"` // 用户注册时间
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"` // 用户信息更新时间
}
