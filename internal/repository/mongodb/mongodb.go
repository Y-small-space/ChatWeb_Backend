package mongodb

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

// 集合名称常量
const (
	UserCollection         = "CHATROOM_DB_users"
	MessageCollection      = "CHATROOM_DB_messages"
	GroupCollection        = "CHATROOM_DB_groups"
	FileCollection         = "CHATROOM_DB_files"
	NotificationCollection = "CHATROOM_DB_notifications"
	FriendshipCollection   = "CHATROOM_DB_friendships"
)

func InitMongoDB(uri, dbName string) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	if err = client.Ping(context.Background(), nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	DB = client.Database(dbName)
	log.Println("Successfully connected to MongoDB")
}

// 获取各个集合的辅助函数
func GetUserCollection() *mongo.Collection {
	return DB.Collection(UserCollection)
}

func GetMessageCollection() *mongo.Collection {
	return DB.Collection(MessageCollection)
}

func GetGroupCollection() *mongo.Collection {
	return DB.Collection(GroupCollection)
}

func GetFileCollection() *mongo.Collection {
	return DB.Collection(FileCollection)
}

func GetNotificationCollection() *mongo.Collection {
	return DB.Collection(NotificationCollection)
}

func GetFriendshipCollection() *mongo.Collection {
	return DB.Collection(FriendshipCollection)
}
