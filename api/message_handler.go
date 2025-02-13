package api

import (
	"chatweb/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// MessageHandler 处理与消息相关的请求，例如标记消息为已读、获取未读消息等
type MessageHandler struct {
	// 服务层，用于处理与消息相关的业务逻辑
	messageService *service.MessageService
}

// NewMessageHandler 构造函数，初始化 MessageHandler
func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// MarkAsRead 标记单条消息为已读
func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	// 获取当前用户的 ID，确保用户已认证
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 从 URL 参数获取消息 ID
	messageID := c.Param("id")
	if messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message ID is required"})
		return
	}

	// 调用服务层标记消息为已读
	if err := h.messageService.MarkMessageAsRead(c.Request.Context(), messageID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回成功消息
	c.JSON(http.StatusOK, gin.H{"message": "Message marked as read"})
}

// MarkMultipleAsRead 标记多条消息为已读
func (h *MessageHandler) MarkMultipleAsRead(c *gin.Context) {
	// 获取当前用户的 ID，确保用户已认证
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 绑定请求数据，获取消息 IDs
	var req struct {
		MessageIDs []string `json:"message_ids" binding:"required"` // 消息 ID 列表
	}

	// 绑定 JSON 数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用服务层标记多条消息为已读
	if err := h.messageService.MarkMessagesAsRead(c.Request.Context(), req.MessageIDs, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回成功消息
	c.JSON(http.StatusOK, gin.H{"message": "Messages marked as read"})
}

// GetUnreadMessages 获取当前用户所有未读的消息
func (h *MessageHandler) GetUnreadMessages(c *gin.Context) {
	// 获取当前用户的 ID，确保用户已认证
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 获取当前用户的未读消息
	messages, err := h.messageService.GetUnreadMessages(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回未读消息
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// GetGroupUnreadMessages 获取指定群组中的未读消息
func (h *MessageHandler) GetGroupUnreadMessages(c *gin.Context) {
	// 获取当前用户的 ID，确保用户已认证
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 从 URL 参数获取群组 ID
	groupID := c.Param("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID is required"})
		return
	}

	// 获取群组中的未读消息
	messages, err := h.messageService.GetGroupUnreadMessages(c.Request.Context(), groupID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取群组未读消息的数量
	unreadCount, err := h.messageService.GetGroupUnreadCount(c.Request.Context(), groupID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回群组未读消息及数量
	c.JSON(http.StatusOK, gin.H{
		"messages":     messages,
		"unread_count": unreadCount,
	})
}

// MarkGroupMessageAsRead 标记群组中的某条消息为已读
func (h *MessageHandler) MarkGroupMessageAsRead(c *gin.Context) {
	// 获取当前用户的 ID，确保用户已认证
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 从 URL 参数获取消息 ID
	messageID := c.Param("id")
	if messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message ID is required"})
		return
	}

	// 调用服务层标记群组消息为已读
	if err := h.messageService.MarkGroupMessageAsRead(c.Request.Context(), messageID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回成功消息
	c.JSON(http.StatusOK, gin.H{"message": "Group message marked as read"})
}
