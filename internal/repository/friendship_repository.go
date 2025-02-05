package repository

import (
	"chatweb/internal/model"
	"chatweb/internal/repository/mongodb"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FriendshipRepository struct {
	collection *mongo.Collection
}

func NewFriendshipRepository() *FriendshipRepository {
	return &FriendshipRepository{
		collection: mongodb.GetFriendshipCollection(),
	}
}

func (r *FriendshipRepository) Create(ctx context.Context, friendship *model.Friendship) error {
	friendship.CreatedAt = time.Now()
	friendship.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, friendship)
	if err != nil {
		return err
	}
	friendship.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *FriendshipRepository) GetFriendsList(ctx context.Context, userID primitive.ObjectID) ([]*model.Friendship, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"user_id": userID},
			{"friend_id": userID},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var friendships []*model.Friendship
	if err := cursor.All(ctx, &friendships); err != nil {
		return nil, err
	}

	return friendships, nil
}
