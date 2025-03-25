package repository

import (
	"chatweb/internal/model"              // 引入数据模型
	"chatweb/internal/repository/mongodb" // 引入 MongoDB 相关的代码
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// FriendshipRepository 是好友关系操作的仓库结构体，包含对好友集合的操作
type FriendshipRepository struct {
	collection *mongo.Collection // MongoDB 中的好友集合
}

// NewFriendshipRepository 返回一个新的 FriendshipRepository 实例，初始化时获取好友集合
func NewFriendshipRepository() *FriendshipRepository {
	return &FriendshipRepository{
		collection: mongodb.GetFriendshipCollection(), // 获取 MongoDB 中的好友集合
	}
}

// Create 创建一个新的好友关系记录，并将好友关系插入到 MongoDB 中
func (r *FriendshipRepository) Create(ctx context.Context, friendship *model.Friendship) error {
	// 设置好友关系的创建时间和更新时间
	friendship.CreatedAt = time.Now()
	friendship.UpdatedAt = time.Now()

	// 将好友关系插入到集合中
	result, err := r.collection.InsertOne(ctx, friendship)
	if err != nil {
		return err // 如果插入失败，返回错误
	}

	// 将插入后返回的好友关系 ID 更新到 friendship 结构体中
	friendship.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetFriendsList 获取指定用户的好友列表
func (r *FriendshipRepository) GetFriendsList(ctx context.Context, userID primitive.ObjectID) ([]*model.Friendship, error) {
	// 创建查询条件，查找用户和好友之间的关系
	filter := bson.M{
		"$or": []bson.M{
			{"user_id": userID},   // 查找用户为 user_id 的记录
			{"friend_id": userID}, // 查找好友为 friend_id 的记录
		},
	}

	// 使用 Find 方法查找符合条件的所有好友关系
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err // 如果查询失败，返回错误
	}
	defer cursor.Close(ctx) // 确保查询游标关闭

	var friendships []*model.Friendship
	// 将查询结果填充到好友关系切片中
	if err := cursor.All(ctx, &friendships); err != nil {
		return nil, err // 如果填充失败，返回错误
	}

	// 返回用户的所有好友关系
	return friendships, nil
}

func (r *FriendshipRepository) Delete(ctx context.Context, userID, friendID primitive.ObjectID) error {
	filter := bson.M{
		"$or": []bson.M{
			{"user_id": userID, "friend_id": friendID},
			{"user_id": friendID, "friend_id": userID},
		},
	}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("friendship not found")
	}

	return nil
}
