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

type FileRepository struct {
	collection *mongo.Collection
}

func NewFileRepository() *FileRepository {
	return &FileRepository{
		collection: mongodb.GetFileCollection(),
	}
}

func (r *FileRepository) Create(ctx context.Context, file *model.File) error {
	file.CreatedAt = time.Now()
	file.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, file)
	if err != nil {
		return err
	}

	file.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *FileRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.File, error) {
	var file model.File
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&file)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*model.File, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var files []*model.File
	if err := cursor.All(ctx, &files); err != nil {
		return nil, err
	}
	return files, nil
}

func (r *FileRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
