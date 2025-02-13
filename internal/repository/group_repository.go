package repository

import (
	"chatweb/internal/model"              // 引入数据模型
	"chatweb/internal/repository/mongodb" // 引入 MongoDB 相关的代码
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GroupRepository 是群组操作的仓库结构体，包含对群组和成员集合的操作
type GroupRepository struct {
	collection       *mongo.Collection // 群组集合
	memberCollection *mongo.Collection // 群组成员集合
}

// NewGroupRepository 返回一个新的 GroupRepository 实例，初始化时获取群组集合和群组成员集合
func NewGroupRepository() *GroupRepository {
	return &GroupRepository{
		collection:       mongodb.GetGroupCollection(),           // 获取 MongoDB 中的群组集合
		memberCollection: mongodb.DB.Collection("group_members"), // 获取 MongoDB 中的群组成员集合
	}
}

// Create 创建一个新的群组记录，并将群组信息插入到 MongoDB 中
func (r *GroupRepository) Create(ctx context.Context, group *model.Group) error {
	// 设置群组的创建时间和更新时间
	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()

	// 将群组插入到群组集合中
	result, err := r.collection.InsertOne(ctx, group)
	if err != nil {
		return err // 如果插入失败，返回错误
	}

	// 将插入后返回的群组 ID 更新到 group 结构体中
	group.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// AddMember 添加群组成员
func (r *GroupRepository) AddMember(ctx context.Context, member *model.GroupMember) error {
	// 设置成员加入时间和更新时间
	member.JoinedAt = time.Now()
	member.UpdatedAt = time.Now()

	// 将群组成员插入到群组成员集合中
	_, err := r.memberCollection.InsertOne(ctx, member)
	return err // 返回操作的错误（如果有）
}

// GetGroupsByUserID 获取指定用户所在的所有群组
func (r *GroupRepository) GetGroupsByUserID(ctx context.Context, userID primitive.ObjectID) ([]*model.Group, error) {
	// 查找用户所在的所有群组成员记录
	cursor, err := r.memberCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err // 如果查询失败，返回错误
	}
	defer cursor.Close(ctx) // 确保查询游标关闭

	var members []model.GroupMember
	// 将查询结果填充到成员切片中
	if err := cursor.All(ctx, &members); err != nil {
		return nil, err // 如果填充失败，返回错误
	}

	// 如果没有成员，返回空的群组列表
	if len(members) == 0 {
		return []*model.Group{}, nil
	}

	// 获取所有群组的 ID
	var groupIDs []primitive.ObjectID
	for _, member := range members {
		groupIDs = append(groupIDs, member.GroupID)
	}

	// 查找群组详情
	groupCursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": groupIDs}})
	if err != nil {
		return nil, err // 如果查询群组失败，返回错误
	}
	defer groupCursor.Close(ctx) // 确保查询游标关闭

	var groups []*model.Group
	// 将查询结果填充到群组切片中
	if err := groupCursor.All(ctx, &groups); err != nil {
		return nil, err // 如果填充失败，返回错误
	}

	// 返回该用户所在的所有群组
	return groups, nil
}

// GetGroupByID 根据群组 ID 查询群组详情
func (r *GroupRepository) GetGroupByID(ctx context.Context, groupID primitive.ObjectID) (*model.Group, error) {
	var group model.Group
	// 使用 FindOne 方法根据群组 ID 查询群组
	err := r.collection.FindOne(ctx, bson.M{"_id": groupID}).Decode(&group)
	if err != nil {
		return nil, err // 如果查询失败，返回错误
	}
	return &group, nil // 返回群组信息
}

// GetGroupMembers 获取指定群组的所有成员
func (r *GroupRepository) GetGroupMembers(ctx context.Context, groupID primitive.ObjectID) ([]*model.GroupMember, error) {
	// 查找群组 ID 对应的所有成员
	cursor, err := r.memberCollection.Find(ctx, bson.M{"group_id": groupID})
	if err != nil {
		return nil, err // 如果查询失败，返回错误
	}
	defer cursor.Close(ctx) // 确保查询游标关闭

	var members []*model.GroupMember
	// 将查询结果填充到成员切片中
	if err := cursor.All(ctx, &members); err != nil {
		return nil, err // 如果填充失败，返回错误
	}

	// 返回群组成员列表
	return members, nil
}

// RemoveMember 移除群组成员
func (r *GroupRepository) RemoveMember(ctx context.Context, groupID, userID primitive.ObjectID) error {
	// 从群组成员集合中删除指定群组和用户的记录
	_, err := r.memberCollection.DeleteOne(ctx, bson.M{
		"group_id": groupID,
		"user_id":  userID,
	})
	return err // 返回删除操作的错误（如果有）
}

// UpdateGroup 更新群组信息
func (r *GroupRepository) UpdateGroup(ctx context.Context, groupID primitive.ObjectID, updates map[string]interface{}) error {
	// 设置更新时间为当前时间
	updates["updated_at"] = time.Now()
	// 更新群组的字段
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": groupID},
		bson.M{"$set": updates}, // 使用 $set 来更新指定字段
	)
	return err // 返回更新操作的错误（如果有）
}
