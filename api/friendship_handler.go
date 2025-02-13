package api

import (
	"chatweb/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// FriendshipHandler 处理好友相关的请求，例如发送好友请求、获取好友列表
type FriendshipHandler struct {
	// 服务层，用于处理与好友相关的业务逻辑
	friendshipService *service.FriendshipService
}

// NewFriendshipHandler 构造函数，初始化 FriendshipHandler
func NewFriendshipHandler(friendshipService *service.FriendshipService) *FriendshipHandler {
	return &FriendshipHandler{
		friendshipService: friendshipService,
	}
}

// SendRequest 处理发送好友请求的操作
func (h *FriendshipHandler) SendRequest(c *gin.Context) {
	// 定义请求参数结构体，绑定请求体中的 JSON 数据
	var req struct {
		FriendID string `json:"user_id" binding:"required"` // 目标好友的用户ID
	}

	// 绑定请求数据，如果出现错误则返回 400 错误
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户的 ID，确保用户已认证
	userID := c.GetString("userID")

	// 调用服务层方法发送好友请求
	if err := h.friendshipService.SendFriendRequest(c.Request.Context(), userID, req.FriendID); err != nil {
		// 如果出错，返回 500 错误和错误信息
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	// 返回成功的消息
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Friend request sent successfully",
	})
}

// GetFriendsList 处理获取当前用户的好友列表的请求
func (h *FriendshipHandler) GetFriendsList(c *gin.Context) {
	// 获取当前用户的 ID，确保用户已认证
	userID := c.GetString("userID")

	// 调用服务层方法获取好友列表
	friends, err := h.friendshipService.GetFriendsList(c.Request.Context(), userID)
	if err != nil {
		// 如果获取好友列表出错，返回 500 错误和错误信息
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回成功的消息和好友列表数据
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"friends": friends,
		},
	})
}
