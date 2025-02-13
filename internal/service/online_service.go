package service

import (
	"chatweb/internal/repository"
	"context"
	"sync"
	"time"

	"chatweb/pkg/event"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OnlineService 处理用户在线状态的服务
type OnlineService struct {
	// userRepo 用户数据仓库，用于更新用户状态
	userRepo *repository.UserRepository
	// onlineUsers 使用 sync.Map 存储在线用户的状态
	onlineUsers sync.Map
	// eventBus 事件总线，用于发布用户上线/下线事件
	eventBus *event.EventBus
}

// NewOnlineService 创建一个新的 OnlineService 实例
func NewOnlineService(userRepo *repository.UserRepository, eventBus *event.EventBus) *OnlineService {
	return &OnlineService{
		userRepo:    userRepo,
		onlineUsers: sync.Map{}, // 使用 sync.Map 处理并发
		eventBus:    eventBus,
	}
}

// SetUserOnline 设置用户为在线状态，并进行相关操作
func (s *OnlineService) SetUserOnline(ctx context.Context, userID string) error {
	// 将用户ID从字符串转为 MongoDB 的 ObjectID 类型
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// 更新数据库中该用户的状态为在线
	updates := map[string]interface{}{
		"online":    true,
		"last_seen": time.Now(), // 更新最后一次在线时间
	}
	if err := s.userRepo.Update(ctx, userObjID, updates); err != nil {
		return err
	}

	// 更新内存缓存，标记用户为在线
	s.onlineUsers.Store(userID, true)

	// 发布用户上线事件
	s.eventBus.Publish(event.Event{
		Type: event.UserOnline, // 事件类型为用户上线
		Content: event.UserStatusContent{
			UserID:   userID,
			IsOnline: true,
		},
	})

	return nil
}

// SetUserOffline 设置用户为离线状态，并进行相关操作
func (s *OnlineService) SetUserOffline(ctx context.Context, userID string) error {
	// 将用户ID从字符串转为 MongoDB 的 ObjectID 类型
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// 更新数据库中该用户的状态为离线
	updates := map[string]interface{}{
		"online":    false,
		"last_seen": time.Now(), // 更新最后一次在线时间
	}
	if err := s.userRepo.Update(ctx, userObjID, updates); err != nil {
		return err
	}

	// 从内存缓存中删除该用户，标记为离线
	s.onlineUsers.Delete(userID)

	// 发布用户离线事件
	s.eventBus.Publish(event.Event{
		Type: event.UserOffline, // 事件类型为用户离线
		Content: event.UserStatusContent{
			UserID:   userID,
			IsOnline: false,
		},
	})

	return nil
}

// IsUserOnline 检查用户是否在线
func (s *OnlineService) IsUserOnline(userID string) bool {
	online, ok := s.onlineUsers.Load(userID) // 从缓存中加载用户状态
	if !ok {
		return false // 如果缓存中没有该用户，则返回离线
	}
	return online.(bool) // 返回用户的在线状态
}

// GetOnlineUsers 获取所有在线的用户ID列表
func (s *OnlineService) GetOnlineUsers(ctx context.Context) ([]string, error) {
	var onlineUsers []string
	// 遍历所有在线的用户，并将其添加到在线用户列表中
	s.onlineUsers.Range(func(key, value interface{}) bool {
		if value.(bool) { // 只有在线的用户才会被添加
			onlineUsers = append(onlineUsers, key.(string))
		}
		return true
	})
	return onlineUsers, nil
}
