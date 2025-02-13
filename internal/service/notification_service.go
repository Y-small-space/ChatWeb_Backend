package service

import (
	"chatweb/internal/model"
	"chatweb/internal/repository"
	"chatweb/pkg/event"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationService 提供了处理通知的业务逻辑
type NotificationService struct {
	notificationRepo *repository.NotificationRepository // 用于操作通知的数据库仓库
	eventBus         *event.EventBus                    // 事件总线，用于发布事件
}

// NewNotificationService 创建一个新的 NotificationService 实例
func NewNotificationService(notificationRepo *repository.NotificationRepository, eventBus *event.EventBus) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		eventBus:         eventBus,
	}
}

// CreateMessageNotification 创建一个新的消息通知
func (s *NotificationService) CreateMessageNotification(ctx context.Context, userID, senderID primitive.ObjectID, message string) error {
	notification := &model.Notification{
		Type:     model.MessageNotification, // 消息类型
		Title:    "新消息",                     // 通知标题
		Content:  message,                   // 通知内容
		UserID:   userID,                    // 接收通知的用户ID
		SenderID: senderID,                  // 发送消息的用户ID
	}

	return s.CreateNotification(ctx, notification) // 调用 CreateNotification 创建通知
}

// CreateGroupNotification 创建一个新的群组通知
func (s *NotificationService) CreateGroupNotification(ctx context.Context, userID, groupID primitive.ObjectID, title, content string) error {
	notification := &model.Notification{
		Type:    model.GroupNotification, // 群组通知类型
		Title:   title,                   // 通知标题
		Content: content,                 // 通知内容
		UserID:  userID,                  // 接收通知的用户ID
		GroupID: groupID,                 // 群组ID
	}

	return s.CreateNotification(ctx, notification) // 调用 CreateNotification 创建通知
}

// CreateSystemNotification 创建一个新的系统通知
func (s *NotificationService) CreateSystemNotification(ctx context.Context, userID primitive.ObjectID, title, content string) error {
	notification := &model.Notification{
		Type:    model.SystemNotification, // 系统通知类型
		Title:   title,                    // 通知标题
		Content: content,                  // 通知内容
		UserID:  userID,                   // 接收通知的用户ID
	}

	return s.CreateNotification(ctx, notification) // 调用 CreateNotification 创建通知
}

// GetUserNotifications 获取用户的通知列表
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID string, limit, offset int) ([]*model.Notification, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID) // 将 userID 转换为 ObjectID
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err) // 错误处理
	}

	return s.notificationRepo.GetUserNotifications(ctx, userObjID, limit, offset) // 获取通知列表
}

// GetUnreadCount 获取用户的未读通知数量
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID) // 将 userID 转换为 ObjectID
	if err != nil {
		return 0, fmt.Errorf("invalid user ID: %v", err) // 错误处理
	}

	return s.notificationRepo.GetUnreadCount(ctx, userObjID) // 获取未读通知数量
}

// MarkAsRead 标记特定通知为已读
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID string) error {
	objID, err := primitive.ObjectIDFromHex(notificationID) // 将通知ID 转换为 ObjectID
	if err != nil {
		return fmt.Errorf("invalid notification ID: %v", err) // 错误处理
	}

	return s.notificationRepo.MarkAsRead(ctx, objID) // 标记通知为已读
}

// MarkAllAsRead 标记用户的所有通知为已读
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	userObjID, err := primitive.ObjectIDFromHex(userID) // 将 userID 转换为 ObjectID
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err) // 错误处理
	}

	return s.notificationRepo.MarkAllAsRead(ctx, userObjID) // 标记所有通知为已读
}

// DeleteNotification 删除特定的通知
func (s *NotificationService) DeleteNotification(ctx context.Context, notificationID string) error {
	objID, err := primitive.ObjectIDFromHex(notificationID) // 将通知ID 转换为 ObjectID
	if err != nil {
		return fmt.Errorf("invalid notification ID: %v", err) // 错误处理
	}

	return s.notificationRepo.Delete(ctx, objID) // 删除通知
}

// DeleteAllNotifications 删除用户的所有通知
func (s *NotificationService) DeleteAllNotifications(ctx context.Context, userID string) error {
	userObjID, err := primitive.ObjectIDFromHex(userID) // 将 userID 转换为 ObjectID
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err) // 错误处理
	}

	return s.notificationRepo.DeleteAllByUserID(ctx, userObjID) // 删除所有通知
}

// CreateNotification 创建一个通用的通知
func (s *NotificationService) CreateNotification(ctx context.Context, notification *model.Notification) error {
	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return err // 如果创建通知失败，则返回错误
	}

	// 发布通知事件
	s.eventBus.Publish(event.Event{
		Type:    event.Notification, // 事件类型
		Content: notification,       // 通知内容
	})

	return nil
}
