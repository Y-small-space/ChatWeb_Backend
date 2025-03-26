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

// MessageService 提供消息相关的操作服务
type MessageService struct {
	messageRepo *repository.MessageRepository // 消息存储库，用于与数据库交互
	readCache   *ReadStatusCache              // 用于存储消息已读状态的缓存
	eventBus    *event.EventBus               // 事件总线，用于发布事件
}

// NewMessageService 创建一个新的 MessageService 实例
func NewMessageService(messageRepo *repository.MessageRepository, readCache *ReadStatusCache, eventBus *event.EventBus) *MessageService {
	return &MessageService{
		messageRepo: messageRepo, // 初始化消息存储库
		readCache:   readCache,   // 初始化已读缓存
		eventBus:    eventBus,    // 初始化事件总线
	}
}

// CreateMessage 创建一条消息
func (s *MessageService) CreateMessage(ctx context.Context, message *model.Message) error {
	log.Println("CreateMessage 创建一条消息")
	log.Printf(message.Content)
	return s.messageRepo.Create(ctx, message) // 将消息存入数据库
}

// GetUserMessages 获取用户与另一个用户的消息
func (s *MessageService) GetMessagesById(ctx context.Context, userID string, otherUserID string) ([]*model.Message, error) {
	// 将用户ID和另一个用户ID转换为 ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	otherUserObjID, err := primitive.ObjectIDFromHex(otherUserID)
	if err != nil {
		return nil, err
	}

	// 设置查询条件，查找两者之间的消息
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
		"group_id": nil, // 排除群组消息
	}

	return s.messageRepo.GetMessages(ctx, filter) // 查询消息
}

// GetAllLastMessages 获取所有用户的最后一条消息
func (s *MessageService) GetAllLastMessages(ctx context.Context, userID string) ([]*model.Message, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	return s.messageRepo.GetAllLastMessages(ctx, userObjID)
}

// GetGroupMessages 获取群组消息
func (s *MessageService) GetGroupMessages(ctx context.Context, groupID string) ([]*model.Message, error) {
	// 将群组ID转换为 ObjectID
	groupObjID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, err
	}

	// 设置查询条件，查找指定群组的消息
	filter := bson.M{"group_id": groupObjID}
	return s.messageRepo.GetMessages(ctx, filter) // 查询群组消息
}

// UpdateMessageStatus 更新消息状态
func (s *MessageService) UpdateMessageStatus(ctx context.Context, messageID string, status string) error {
	// 将消息ID转换为 ObjectID
	objID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return err
	}

	return s.messageRepo.UpdateStatus(ctx, objID, status) // 更新消息状态
}

// MarkMessageAsRead 标记单条消息为已读
func (s *MessageService) MarkMessageAsRead(ctx context.Context, messageID string, userID string) error {
	// 将消息ID和用户ID转换为 ObjectID
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

	return s.messageRepo.MarkAsRead(ctx, msgObjID, userObjID) // 更新消息为已读
}

// MarkMessagesAsRead 批量标记消息为已读
func (s *MessageService) MarkMessagesAsRead(ctx context.Context, messageIDs []string, userID string) error {
	var msgObjIDs []primitive.ObjectID
	// 将消息ID数组转换为 ObjectID
	for _, id := range messageIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return fmt.Errorf("invalid message ID %s: %v", id, err)
		}
		msgObjIDs = append(msgObjIDs, objID)
	}

	// 将用户ID转换为 ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	return s.messageRepo.MarkMultipleAsRead(ctx, msgObjIDs, userObjID) // 批量更新消息为已读
}

// GetUnreadMessages 获取用户未读的消息
func (s *MessageService) GetUnreadMessages(ctx context.Context, userID string) ([]*model.Message, error) {
	// 将用户ID转换为 ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	return s.messageRepo.GetUnreadMessages(ctx, userObjID) // 查询未读消息
}

// GetGroupUnreadMessages 获取群组中用户未读的消息
func (s *MessageService) GetGroupUnreadMessages(ctx context.Context, groupID string, userID string) ([]*model.Message, error) {
	// 将群组ID和用户ID转换为 ObjectID
	groupObjID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID: %v", err)
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	return s.messageRepo.GetGroupUnreadMessages(ctx, groupObjID, userObjID) // 查询群组未读消息
}

// GetGroupUnreadCount 获取群组中用户未读消息的数量
func (s *MessageService) GetGroupUnreadCount(ctx context.Context, groupID string, userID string) (int64, error) {
	// 将群组ID和用户ID转换为 ObjectID
	groupObjID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return 0, fmt.Errorf("invalid group ID: %v", err)
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID: %v", err)
	}

	return s.messageRepo.GetGroupUnreadCount(ctx, groupObjID, userObjID) // 获取群组未读消息数量
}

// GetMessageReadStatus 获取消息的已读状态
func (s *MessageService) GetMessageReadStatus(ctx context.Context, messageID string) (*ReadStatus, error) {
	status, err := s.readCache.GetReadStatus(ctx, messageID)
	if err == nil {
		return status, nil // 如果缓存中存在，直接返回已读状态
	}

	// 如果缓存中没有，查询数据库
	msgObjID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return nil, fmt.Errorf("invalid message ID: %v", err)
	}

	message, err := s.messageRepo.GetByID(ctx, msgObjID)
	if err != nil {
		return nil, err
	}

	// 获取最后一条已读记录
	if len(message.ReadBy) == 0 {
		return nil, nil // 没有已读记录
	}

	lastRead := message.ReadBy[len(message.ReadBy)-1]
	status = &ReadStatus{
		MessageID: messageID,
		UserID:    lastRead.UserID.Hex(),
		ReadAt:    lastRead.ReadAt,
	}

	// 将已读状态缓存
	if err := s.readCache.SetReadStatus(ctx, status); err != nil {
		log.Printf("Failed to cache read status: %v", err)
	}

	return status, nil
}

// MarkGroupMessageAsRead 标记群组消息为已读
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
