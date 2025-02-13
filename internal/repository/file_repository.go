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

// FileRepository 是文件操作的仓库结构体，包含对文件集合的操作
type FileRepository struct {
	collection *mongo.Collection // MongoDB 中的文件集合
}

// NewFileRepository 返回一个新的 FileRepository 实例，初始化时获取文件集合
func NewFileRepository() *FileRepository {
	return &FileRepository{
		collection: mongodb.GetFileCollection(), // 获取 MongoDB 中的文件集合
	}
}

// Create 创建一个新的文件记录，并将文件信息插入到 MongoDB 中
func (r *FileRepository) Create(ctx context.Context, file *model.File) error {
	// 设置文件的创建时间和更新时间
	file.CreatedAt = time.Now()
	file.UpdatedAt = time.Now()

	// 将文件插入到集合中
	result, err := r.collection.InsertOne(ctx, file)
	if err != nil {
		return err // 如果插入失败，返回错误
	}

	// 将插入后返回的文件 ID 更新到 file 结构体中
	file.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID 根据文件的 ID 查询文件记录
func (r *FileRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.File, error) {
	var file model.File
	// 使用 FindOne 方法按 ID 查找文件
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&file)
	if err != nil {
		return nil, err // 如果查询失败，返回错误
	}
	return &file, nil // 返回文件记录
}

// GetByUserID 根据用户 ID 查询该用户的所有文件
func (r *FileRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*model.File, error) {
	// 查找所有与指定用户 ID 相关的文件
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err // 如果查询失败，返回错误
	}
	defer cursor.Close(ctx) // 确保游标关闭

	var files []*model.File
	// 将查询结果填充到文件切片中
	if err := cursor.All(ctx, &files); err != nil {
		return nil, err // 如果填充失败，返回错误
	}
	return files, nil // 返回文件列表
}

// Delete 删除指定 ID 的文件记录
func (r *FileRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	// 根据文件 ID 删除记录
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err // 返回删除操作的错误（如果有）
}
