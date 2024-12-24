package service

import (
	"chatweb/pkg/cache"
	"context"
	"fmt"
	"time"
)

const (
	readStatusKeyPrefix = "read_status:"
	readStatusTTL       = 24 * time.Hour
)

type ReadStatusCache struct {
	redis *cache.RedisClient
}

func NewReadStatusCache(redis *cache.RedisClient) *ReadStatusCache {
	return &ReadStatusCache{
		redis: redis,
	}
}

type ReadStatus struct {
	MessageID string    `json:"message_id"`
	UserID    string    `json:"user_id"`
	ReadAt    time.Time `json:"read_at"`
}

func (c *ReadStatusCache) getCacheKey(messageID string) string {
	return fmt.Sprintf("%s%s", readStatusKeyPrefix, messageID)
}

func (c *ReadStatusCache) SetReadStatus(ctx context.Context, status *ReadStatus) error {
	key := c.getCacheKey(status.MessageID)
	return c.redis.Set(ctx, key, status, readStatusTTL)
}

func (c *ReadStatusCache) GetReadStatus(ctx context.Context, messageID string) (*ReadStatus, error) {
	key := c.getCacheKey(messageID)
	var status ReadStatus
	err := c.redis.Get(ctx, key, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (c *ReadStatusCache) DeleteReadStatus(ctx context.Context, messageID string) error {
	key := c.getCacheKey(messageID)
	return c.redis.Delete(ctx, key)
}
