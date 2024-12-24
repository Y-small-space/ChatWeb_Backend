package repository

import (
	"chatweb/internal/model"
	"chatweb/internal/repository/mongodb"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NotificationRepository struct {
	collection *mongo.Collection
}

func NewNotificationRepository() *NotificationRepository {
	return &NotificationRepository{
		collection: mongodb.GetNotificationCollection(),
	}
}

func (r *NotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()
	notification.IsRead = false

	result, err := r.collection.InsertOne(ctx, notification)
	if err != nil {
		return err
	}

	notification.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *NotificationRepository) GetUserNotifications(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*model.Notification, error) {
	opts := options.Find().
		SetSort(bson.D{primitive.E{Key: "created_at", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*model.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"user_id": userID,
		"is_read": false,
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, notificationID primitive.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"is_read":    true,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": notificationID}, update)
	return err
}

func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID primitive.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"is_read":    true,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateMany(ctx, bson.M{
		"user_id": userID,
		"is_read": false,
	}, update)
	return err
}

func (r *NotificationRepository) Delete(ctx context.Context, notificationID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": notificationID})
	return err
}

func (r *NotificationRepository) DeleteAllByUserID(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}
