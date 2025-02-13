package api

import (
	"chatweb/internal/service" // 引入服务层
	"net/http"
	"strconv" // 用于字符串转数字

	"github.com/gin-gonic/gin" // Gin框架
)

type NotificationHandler struct {
	notificationService *service.NotificationService // 引入通知服务
}

// NewNotificationHandler：构造函数，用于初始化 NotificationHandler
func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService, // 注入通知服务
	}
}

// GetNotifications：获取用户的通知列表，支持分页
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	// 获取用户ID，通常从JWT中获取
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"}) // 如果没有获取到用户ID，返回未授权
		return
	}

	// 获取分页参数，默认每页20条，偏移量默认0
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))  // 转换limit为整数
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0")) // 转换offset为整数

	// 调用服务层获取通知列表
	notifications, err := h.notificationService.GetUserNotifications(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // 错误处理
		return
	}

	// 获取未读通知的数量
	unreadCount, err := h.notificationService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // 错误处理
		return
	}

	// 返回通知列表和未读通知数量
	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications, // 通知列表
		"unread_count":  unreadCount,   // 未读通知数量
	})
}

// MarkAsRead：标记单个通知为已读
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	// 获取用户ID
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"}) // 未授权错误
		return
	}

	// 从路径参数获取通知ID
	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notification ID is required"}) // 如果没有提供通知ID，返回BadRequest错误
		return
	}

	// 调用服务层将通知标记为已读
	if err := h.notificationService.MarkAsRead(c.Request.Context(), notificationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // 错误处理
		return
	}

	// 返回成功标记为已读的响应
	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllAsRead：标记所有通知为已读
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	// 获取用户ID
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"}) // 未授权错误
		return
	}

	// 调用服务层将所有通知标记为已读
	if err := h.notificationService.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // 错误处理
		return
	}

	// 返回成功标记所有通知为已读的响应
	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// DeleteNotification：删除单个通知
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	// 获取用户ID
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"}) // 未授权错误
		return
	}

	// 从路径参数获取通知ID
	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notification ID is required"}) // 如果没有提供通知ID，返回BadRequest错误
		return
	}

	// 调用服务层删除通知
	if err := h.notificationService.DeleteNotification(c.Request.Context(), notificationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // 错误处理
		return
	}

	// 返回成功删除通知的响应
	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted"})
}

// DeleteAllNotifications：删除所有通知
func (h *NotificationHandler) DeleteAllNotifications(c *gin.Context) {
	// 获取用户ID
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"}) // 未授权错误
		return
	}

	// 调用服务层删除所有通知
	if err := h.notificationService.DeleteAllNotifications(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // 错误处理
		return
	}

	// 返回成功删除所有通知的响应
	c.JSON(http.StatusOK, gin.H{"message": "All notifications deleted"})
}
