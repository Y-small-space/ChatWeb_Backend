package service

import (
	"chatweb/internal/model"
	"chatweb/internal/repository"
	"context"
	"fmt"
	"log"
	"time"

	"chatweb/pkg/event"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MessageService struct {
	messageRepo *repository.MessageRepository
	readCache   *ReadStatusCache
	eventBus    *event.EventBus
}

func NewMessageService(messageRepo *repository.MessageRepository, readCache *ReadStatusCache, eventBus *event.EventBus) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		readCache:   readCache,
		eventBus:    eventBus,
	}
}

func (s *MessageService) CreateMessage(ctx context.Context, message *model.Message) error {
	return s.messageRepo.Create(ctx, message)
}

func (s *MessageService) GetUserMessages(ctx context.Context, userID string, otherUserID string, limit, offset int) ([]*model.Message, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	otherUserObjID, err := primitive.ObjectIDFromHex(otherUserID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"$or": []bson.M{
			{
				"sender_id":   userObjID,
				"receiver_id": otherUserObjID,
			},
			{
				"sender_id":   otherUserObjID,
				"receiver_id": userObjID,
			},
		},
		"group_id": nil,
	}

	return s.messageRepo.GetMessages(ctx, filter, limit, offset)
}

func (s *MessageService) GetGroupMessages(ctx context.Context, groupID string, limit, offset int) ([]*model.Message, error) {
	groupObjID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"group_id": groupObjID}
	return s.messageRepo.GetMessages(ctx, filter, limit, offset)
}

func (s *MessageService) UpdateMessageStatus(ctx context.Context, messageID string, status string) error {
	objID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return err
	}

	return s.messageRepo.UpdateStatus(ctx, objID, status)
}

func (s *MessageService) MarkMessageAsRead(ctx context.Context, messageID string, userID string) error {
	msgObjID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %v", err)
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	// 发布消息已读事件
	s.eventBus.Publish(event.Event{
		Type: event.MessageRead,
		Content: event.MessageReadContent{
			MessageID: messageID,
			UserID:    userID,
			ReadAt:    time.Now().Format(time.RFC3339),
			IsGroup:   false,
		},
	})

	return s.messageRepo.MarkAsRead(ctx, msgObjID, userObjID)
}

func (s *MessageService) MarkMessagesAsRead(ctx context.Context, messageIDs []string, userID string) error {
	var msgObjIDs []primitive.ObjectID
	for _, id := range messageIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return fmt.Errorf("invalid message ID %s: %v", id, err)
		}
		msgObjIDs = append(msgObjIDs, objID)
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	return s.messageRepo.MarkMultipleAsRead(ctx, msgObjIDs, userObjID)
}

func (s *MessageService) GetUnreadMessages(ctx context.Context, userID string) ([]*model.Message, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	return s.messageRepo.GetUnreadMessages(ctx, userObjID)
}

func (s *MessageService) GetGroupUnreadMessages(ctx context.Context, groupID string, userID string) ([]*model.Message, error) {
	groupObjID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID: %v", err)
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	return s.messageRepo.GetGroupUnreadMessages(ctx, groupObjID, userObjID)
}

func (s *MessageService) GetGroupUnreadCount(ctx context.Context, groupID string, userID string) (int64, error) {
	groupObjID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return 0, fmt.Errorf("invalid group ID: %v", err)
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID: %v", err)
	}

	return s.messageRepo.GetGroupUnreadCount(ctx, groupObjID, userObjID)
}

func (s *MessageService) GetMessageReadStatus(ctx context.Context, messageID string) (*ReadStatus, error) {
	status, err := s.readCache.GetReadStatus(ctx, messageID)
	if err == nil {
		return status, nil
	}

	msgObjID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return nil, fmt.Errorf("invalid message ID: %v", err)
	}

	message, err := s.messageRepo.GetByID(ctx, msgObjID)
	if err != nil {
		return nil, err
	}

	if len(message.ReadBy) == 0 {
		return nil, nil
	}

	lastRead := message.ReadBy[len(message.ReadBy)-1]
	status = &ReadStatus{
		MessageID: messageID,
		UserID:    lastRead.UserID.Hex(),
		ReadAt:    lastRead.ReadAt,
	}

	if err := s.readCache.SetReadStatus(ctx, status); err != nil {
		log.Printf("Failed to cache read status: %v", err)
	}

	return status, nil
}

func (s *MessageService) MarkGroupMessageAsRead(ctx context.Context, messageID string, userID string) error {
	msgObjID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %v", err)
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	// 获取消息信息
	message, err := s.messageRepo.GetByID(ctx, msgObjID)
	if err != nil {
		return err
	}

	// 检查是否是群组消息
	if message.GroupID.IsZero() {
		return fmt.Errorf("not a group message")
	}

	// 更新已读状态
	if err := s.messageRepo.MarkAsRead(ctx, msgObjID, userObjID); err != nil {
		return err
	}

	// 获取已读用户列表
	readBy := make([]string, len(message.ReadBy)+1)
	for i, receipt := range message.ReadBy {
		readBy[i] = receipt.UserID.Hex()
	}
	readBy[len(readBy)-1] = userID

	// 发布群组消息已读事件
	s.eventBus.Publish(event.Event{
		Type: event.GroupRead,
		Content: event.GroupReadContent{
			MessageID:  messageID,
			GroupID:    message.GroupID.Hex(),
			ReadByUser: userID,
			ReadAt:     time.Now().Format(time.RFC3339),
			ReadCount:  len(readBy),
			ReadBy:     readBy,
		},
	})

	return nil
}
