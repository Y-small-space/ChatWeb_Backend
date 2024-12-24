package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileType string

const (
	ImageFile FileType = "image"
	VideoFile FileType = "video"
	AudioFile FileType = "audio"
	DocFile   FileType = "document"
)

type File struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Type      FileType           `bson:"type" json:"type"`
	Size      int64              `bson:"size" json:"size"`
	URL       string             `bson:"url" json:"url"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
