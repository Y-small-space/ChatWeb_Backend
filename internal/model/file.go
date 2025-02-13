package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileType 定义文件类型（图片、视频、音频、文档）
type FileType string

const (
	ImageFile FileType = "image"    // 图片
	VideoFile FileType = "video"    // 视频
	AudioFile FileType = "audio"    // 音频
	DocFile   FileType = "document" // 文档
)

// File 定义文件数据结构
type File struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`      // 文件的唯一标识符
	Name      string             `bson:"name" json:"name"`             // 文件名
	Type      FileType           `bson:"type" json:"type"`             // 文件类型（图片、视频、音频、文档）
	Size      int64              `bson:"size" json:"size"`             // 文件大小（字节）
	URL       string             `bson:"url" json:"url"`               // 文件存储的 URL 地址
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`       // 上传文件的用户 ID
	CreatedAt time.Time          `bson:"created_at" json:"created_at"` // 文件上传时间
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"` // 文件更新时间
}
