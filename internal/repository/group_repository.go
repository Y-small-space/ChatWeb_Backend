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

type GroupRepository struct {
	collection       *mongo.Collection
	memberCollection *mongo.Collection
}

func NewGroupRepository() *GroupRepository {
	return &GroupRepository{
		collection:       mongodb.GetGroupCollection(),
		memberCollection: mongodb.DB.Collection("group_members"),
	}
}

func (r *GroupRepository) Create(ctx context.Context, group *model.Group) error {
	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, group)
	if err != nil {
		return err
	}

	group.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *GroupRepository) AddMember(ctx context.Context, member *model.GroupMember) error {
	member.JoinedAt = time.Now()
	member.UpdatedAt = time.Now()

	_, err := r.memberCollection.InsertOne(ctx, member)
	return err
}

func (r *GroupRepository) GetGroupsByUserID(ctx context.Context, userID primitive.ObjectID) ([]*model.Group, error) {
	// 查找用户所在的所有群组
	cursor, err := r.memberCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []model.GroupMember
	if err := cursor.All(ctx, &members); err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return []*model.Group{}, nil
	}

	// 获取群组ID列表
	var groupIDs []primitive.ObjectID
	for _, member := range members {
		groupIDs = append(groupIDs, member.GroupID)
	}

	// 查找群组详情
	groupCursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": groupIDs}})
	if err != nil {
		return nil, err
	}
	defer groupCursor.Close(ctx)

	var groups []*model.Group
	if err := groupCursor.All(ctx, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func (r *GroupRepository) GetGroupByID(ctx context.Context, groupID primitive.ObjectID) (*model.Group, error) {
	var group model.Group
	err := r.collection.FindOne(ctx, bson.M{"_id": groupID}).Decode(&group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepository) GetGroupMembers(ctx context.Context, groupID primitive.ObjectID) ([]*model.GroupMember, error) {
	cursor, err := r.memberCollection.Find(ctx, bson.M{"group_id": groupID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []*model.GroupMember
	if err := cursor.All(ctx, &members); err != nil {
		return nil, err
	}

	return members, nil
}

func (r *GroupRepository) RemoveMember(ctx context.Context, groupID, userID primitive.ObjectID) error {
	_, err := r.memberCollection.DeleteOne(ctx, bson.M{
		"group_id": groupID,
		"user_id":  userID,
	})
	return err
}

func (r *GroupRepository) UpdateGroup(ctx context.Context, groupID primitive.ObjectID, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": groupID},
		bson.M{"$set": updates},
	)
	return err
}
