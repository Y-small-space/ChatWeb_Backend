package mongodb

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB 是全局的 MongoDB 数据库实例
var DB *mongo.Database

// 定义一些常量，表示数据库中各个集合的名称
const (
	UserCollection         = "CHATROOM_DB_users"         // 用户集合
	MessageCollection      = "CHATROOM_DB_messages"      // 消息集合
	GroupCollection        = "CHATROOM_DB_groups"        // 群组集合
	FileCollection         = "CHATROOM_DB_files"         // 文件集合
	NotificationCollection = "CHATROOM_DB_notifications" // 通知集合
	FriendshipCollection   = "CHATROOM_DB_friendships"   // 好友关系集合
)

// InitMongoDB 用于初始化 MongoDB 连接
func InitMongoDB(uri, dbName string) {
	// 使用给定的 URI 创建 MongoDB 客户端
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err) // 如果连接失败，日志并退出
	}

	// 检查 MongoDB 是否正常连接
	if err = client.Ping(context.Background(), nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err) // 如果连接失败，日志并退出
	}

	// 获取指定数据库
	DB = client.Database(dbName)
	log.Println("Successfully connected to MongoDB") // 连接成功，输出日志
}

// 以下是一些辅助函数，用于获取 MongoDB 数据库中的集合
// 这些函数会返回对应集合的 *mongo.Collection 对象，可以用来执行数据库操作

// GetUserCollection 获取用户集合
func GetUserCollection() *mongo.Collection {
	return DB.Collection(UserCollection)
}

// GetMessageCollection 获取消息集合
func GetMessageCollection() *mongo.Collection {
	return DB.Collection(MessageCollection)
}

// GetGroupCollection 获取群组集合
func GetGroupCollection() *mongo.Collection {
	return DB.Collection(GroupCollection)
}

// GetFileCollection 获取文件集合
func GetFileCollection() *mongo.Collection {
	return DB.Collection(FileCollection)
}

// GetNotificationCollection 获取通知集合
func GetNotificationCollection() *mongo.Collection {
	return DB.Collection(NotificationCollection)
}

// GetFriendshipCollection 获取好友关系集合
func GetFriendshipCollection() *mongo.Collection {
	return DB.Collection(FriendshipCollection)
}
