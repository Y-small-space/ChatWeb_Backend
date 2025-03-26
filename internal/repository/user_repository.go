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
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		collection: mongodb.GetUserCollection(),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"phone": phone}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	log.Print("findByiD", id)
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, id primitive.ObjectID, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	return err
}

// SearchUserByIdentifier 通过标识符（手机号、邮箱或用户名）搜索用户
func (r *UserRepository) SearchUserByIdentifier(ctx context.Context, identifier string) (*model.User, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"email": identifier},
			{"username": identifier},
			{"phone": identifier},
		},
	}

	var user model.User
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// FindByIDs 批量查询用户信息
func (r *UserRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
	var users []*model.User

	log.Print("ids", ids)
	// 查询条件：用户 ID 在给定的 ID 列表中
	filter := bson.M{"_id": bson.M{"$in": ids}}

	// 执行数据库查询
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 解析查询结果
	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			continue
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUserAvatar 更新用户头像 URL
func (r *UserRepository) UpdateUserAvatar(ctx context.Context, userID string, avatarURL string) error {
	log.Print("userId", userID)
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Print("err", err)
		return err
	}

	log.Print("objID", objID)
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"avatar": avatarURL, "updated_at": time.Now()}}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}
