package service

import (
	"chatweb/internal/repository"
	"context"
	"sync"
	"time"

	"chatweb/pkg/event"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OnlineService struct {
	userRepo    *repository.UserRepository
	onlineUsers sync.Map
	eventBus    *event.EventBus
}

func NewOnlineService(userRepo *repository.UserRepository, eventBus *event.EventBus) *OnlineService {
	return &OnlineService{
		userRepo:    userRepo,
		onlineUsers: sync.Map{},
		eventBus:    eventBus,
	}
}

func (s *OnlineService) SetUserOnline(ctx context.Context, userID string) error {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// 更新数据库中的用户状态
	updates := map[string]interface{}{
		"online":    true,
		"last_seen": time.Now(),
	}
	if err := s.userRepo.Update(ctx, userObjID, updates); err != nil {
		return err
	}

	// 更新内存缓存
	s.onlineUsers.Store(userID, true)

	// 发布用户在线事件
	s.eventBus.Publish(event.Event{
		Type: event.UserOnline,
		Content: event.UserStatusContent{
			UserID:   userID,
			IsOnline: true,
		},
	})

	return nil
}

func (s *OnlineService) SetUserOffline(ctx context.Context, userID string) error {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// 更新数据库中的用户状态
	updates := map[string]interface{}{
		"online":    false,
		"last_seen": time.Now(),
	}
	if err := s.userRepo.Update(ctx, userObjID, updates); err != nil {
		return err
	}

	// 更新内存缓存
	s.onlineUsers.Delete(userID)

	// 发布用户离线事件
	s.eventBus.Publish(event.Event{
		Type: event.UserOffline,
		Content: event.UserStatusContent{
			UserID:   userID,
			IsOnline: false,
		},
	})

	return nil
}

func (s *OnlineService) IsUserOnline(userID string) bool {
	online, ok := s.onlineUsers.Load(userID)
	if !ok {
		return false
	}
	return online.(bool)
}

func (s *OnlineService) GetOnlineUsers(ctx context.Context) ([]string, error) {
	var onlineUsers []string
	s.onlineUsers.Range(func(key, value interface{}) bool {
		if value.(bool) {
			onlineUsers = append(onlineUsers, key.(string))
		}
		return true
	})
	return onlineUsers, nil
}
