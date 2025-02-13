package repository

import (
	"chatweb/internal/model"
	"chatweb/internal/repository/mongodb"
	"context"
	"fmt"
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
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, message)
	if err != nil {
		return err
	}

	message.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *MessageRepository) GetMessages(ctx context.Context, filter bson.M, limit, offset int) ([]*model.Message, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

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
