package service

import (
	"chatweb/pkg/cache"
	"context"
	"fmt"
	"time"
)

// 常量定义
const (
	readStatusKeyPrefix = "read_status:" // Redis键的前缀
	readStatusTTL       = 24 * time.Hour // 缓存有效期，设为24小时
)

// ReadStatusCache 提供了管理消息已读状态的缓存服务
type ReadStatusCache struct {
	redis *cache.RedisClient // Redis客户端，用于与Redis交互
}

// NewReadStatusCache 创建一个新的 ReadStatusCache 实例
func NewReadStatusCache(redis *cache.RedisClient) *ReadStatusCache {
	return &ReadStatusCache{
		redis: redis, // 初始化 Redis 客户端
	}
}

// ReadStatus 表示消息的已读状态
type ReadStatus struct {
	MessageID string    `json:"message_id"` // 消息ID
	UserID    string    `json:"user_id"`    // 用户ID
	ReadAt    time.Time `json:"read_at"`    // 已读时间
}

// getCacheKey 生成缓存的键（即Redis的键）
func (c *ReadStatusCache) getCacheKey(messageID string) string {
	// 使用消息ID和预定义的前缀来生成Redis键
	return fmt.Sprintf("%s%s", readStatusKeyPrefix, messageID)
}

// SetReadStatus 设置某个消息的已读状态
func (c *ReadStatusCache) SetReadStatus(ctx context.Context, status *ReadStatus) error {
	// 生成缓存键
	key := c.getCacheKey(status.MessageID)
	// 将已读状态存入Redis，设置过期时间
	return c.redis.Set(ctx, key, status, readStatusTTL)
}

// GetReadStatus 获取某个消息的已读状态
func (c *ReadStatusCache) GetReadStatus(ctx context.Context, messageID string) (*ReadStatus, error) {
	// 生成缓存键
	key := c.getCacheKey(messageID)
	// 从Redis获取已读状态
	var status ReadStatus
	err := c.redis.Get(ctx, key, &status)
	if err != nil {
		return nil, err // 如果读取失败，返回错误
	}
	// 返回已读状态
	return &status, nil
}

// DeleteReadStatus 删除某个消息的已读状态
func (c *ReadStatusCache) DeleteReadStatus(ctx context.Context, messageID string) error {
	// 生成缓存键
	key := c.getCacheKey(messageID)
	// 删除Redis中的已读状态
	return c.redis.Delete(ctx, key)
}
