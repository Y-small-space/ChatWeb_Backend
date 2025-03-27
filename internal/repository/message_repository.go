package repository

import (
	"chatweb/internal/model"
	"chatweb/internal/repository/mongodb"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageRepository struct {
	collection *mongo.Collection
}

func NewMessageRepository() *MessageRepository {
	return &MessageRepository{
		collection: mongodb.GetMessageCollection(),
	}
}

func (r *MessageRepository) Create(ctx context.Context, message *model.Message) error {
	log.Println("createMessage", message)
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, message)
	log.Println(result)
	if err != nil {
		return err
	}

	message.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *MessageRepository) GetMessages(ctx context.Context, filter bson.M) ([]*model.Message, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*model.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepository) GetAllLastMessages(ctx context.Context, userId primitive.ObjectID) ([]*model.Message, error) {
	var messages []*model.Message

	// 1. 过滤当前用户相关的消息
	matchStage := bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "sender_id", Value: userId}},
				bson.D{{Key: "receiver_id", Value: userId}},
			}},
		}},
	}

	// 2. 按聊天对象分组，取最新一条消息
	groupStage := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "$cond", Value: bson.D{
					{Key: "if", Value: bson.D{{Key: "$eq", Value: bson.A{"$sender_id", userId}}}},
					{Key: "then", Value: "$receiver_id"},
					{Key: "else", Value: "$sender_id"},
				}},
			}},
			{Key: "lastMessage", Value: bson.D{{Key: "$last", Value: "$$ROOT"}}}, // 取最新一条
		}},
	}

	// 3. 只保留 lastMessage 字段
	projectStage := bson.D{
		{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$lastMessage"}}},
	}

	// 4. 按时间倒序排序
	sortStage := bson.D{
		{Key: "$sort", Value: bson.D{{"created_at", -1}}},
	}

	// 执行聚合查询
	cursor, err := r.collection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage, sortStage})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 解析查询结果
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// DeleteMessageById 根据 userId, otherId 和 messageId 删除特定的消息
func (r *MessageRepository) DeleteMessageById(ctx context.Context, filter bson.M) error {
	// 删除消息
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	// 如果没有删除任何记录，返回错误
	if result.DeletedCount == 0 {
		return fmt.Errorf("message not found or you do not have permission to delete it")
	}

	return nil
}

func (r *MessageRepository) UpdateStatus(ctx context.Context, messageID primitive.ObjectID, status string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": messageID}, update)
	return err
}

func (r *MessageRepository) MarkAsRead(ctx context.Context, messageID primitive.ObjectID, userID primitive.ObjectID) error {
	readReceipt := model.ReadReceipt{
		UserID:    userID,
		ReadAt:    time.Now(),
		Timestamp: time.Now(),
	}

	update := bson.M{
		"$push": bson.M{
			"read_by": readReceipt,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": messageID}, update)
	return err
}

func (r *MessageRepository) MarkMultipleAsRead(ctx context.Context, messageIDs []primitive.ObjectID, userID primitive.ObjectID) error {
	readReceipt := model.ReadReceipt{
		UserID:    userID,
		ReadAt:    time.Now(),
		Timestamp: time.Now(),
	}

	update := bson.M{
		"$push": bson.M{
			"read_by": readReceipt,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateMany(ctx, bson.M{
		"_id": bson.M{"$in": messageIDs},
		"read_by.user_id": bson.M{
			"$ne": userID,
		},
	}, update)
	return err
}

func (r *MessageRepository) GetUnreadMessages(ctx context.Context, userID primitive.ObjectID) ([]*model.Message, error) {
	filter := bson.M{
		"receiver_id": userID,
		"read_by.user_id": bson.M{
			"$ne": userID,
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*model.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepository) GetGroupUnreadMessages(ctx context.Context, groupID primitive.ObjectID, userID primitive.ObjectID) ([]*model.Message, error) {
	filter := bson.M{
		"group_id": groupID,
		"read_by.user_id": bson.M{
			"$ne": userID,
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*model.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepository) GetGroupUnreadCount(ctx context.Context, groupID primitive.ObjectID, userID primitive.ObjectID) (int64, error) {
	filter := bson.M{
		"group_id": groupID,
		"read_by.user_id": bson.M{
			"$ne": userID,
		},
	}

	return r.collection.CountDocuments(ctx, filter)
}

func (r *MessageRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.Message, error) {
	var message model.Message
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&message)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("message not found")
		}
		return nil, err
	}
	return &message, nil
}
