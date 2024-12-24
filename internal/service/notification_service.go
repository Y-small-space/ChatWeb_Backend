package service

import (
	"chatweb/internal/model"
	"chatweb/internal/repository"
	"chatweb/pkg/event"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	eventBus         *event.EventBus
}

func NewNotificationService(notificationRepo *repository.NotificationRepository, eventBus *event.EventBus) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		eventBus:         eventBus,
	}
}

func (s *NotificationService) CreateMessageNotification(ctx context.Context, userID, senderID primitive.ObjectID, message string) error {
	notification := &model.Notification{
		Type:     model.MessageNotification,
		Title:    "新消息",
		Content:  message,
		UserID:   userID,
		SenderID: senderID,
	}

	return s.CreateNotification(ctx, notification)
}

func (s *NotificationService) CreateGroupNotification(ctx context.Context, userID, groupID primitive.ObjectID, title, content string) error {
	notification := &model.Notification{
		Type:    model.GroupNotification,
		Title:   title,
		Content: content,
		UserID:  userID,
		GroupID: groupID,
	}

	return s.CreateNotification(ctx, notification)
}

func (s *NotificationService) CreateSystemNotification(ctx context.Context, userID primitive.ObjectID, title, content string) error {
	notification := &model.Notification{
		Type:    model.SystemNotification,
		Title:   title,
		Content: content,
		UserID:  userID,
	}

	return s.CreateNotification(ctx, notification)
}

func (s *NotificationService) GetUserNotifications(ctx context.Context, userID string, limit, offset int) ([]*model.Notification, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	return s.notificationRepo.GetUserNotifications(ctx, userObjID, limit, offset)
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID: %v", err)
	}

	return s.notificationRepo.GetUnreadCount(ctx, userObjID)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID string) error {
	objID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return fmt.Errorf("invalid notification ID: %v", err)
	}

	return s.notificationRepo.MarkAsRead(ctx, objID)
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	return s.notificationRepo.MarkAllAsRead(ctx, userObjID)
}

func (s *NotificationService) DeleteNotification(ctx context.Context, notificationID string) error {
	objID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return fmt.Errorf("invalid notification ID: %v", err)
	}

	return s.notificationRepo.Delete(ctx, objID)
}

func (s *NotificationService) DeleteAllNotifications(ctx context.Context, userID string) error {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	return s.notificationRepo.DeleteAllByUserID(ctx, userObjID)
}

func (s *NotificationService) CreateNotification(ctx context.Context, notification *model.Notification) error {
	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err
	}

	// 发布通知事件
	s.eventBus.Publish(event.Event{
		Type:    event.Notification,
		Content: notification,
	})

	return nil
}
