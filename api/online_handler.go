package api

import (
	"chatweb/internal/service" // 引入服务层
	"net/http"                 // HTTP 状态码

	"github.com/gin-gonic/gin" // Gin 框架
)

// OnlineHandler：处理与在线用户相关的 API 请求
type OnlineHandler struct {
	onlineService *service.OnlineService // 引入在线状态服务
}

// NewOnlineHandler：构造函数，用于初始化 OnlineHandler
func NewOnlineHandler(onlineService *service.OnlineService) *OnlineHandler {
	return &OnlineHandler{
		onlineService: onlineService, // 注入在线状态服务
	}
}

// GetOnlineUsers：获取当前在线的用户列表
func (h *OnlineHandler) GetOnlineUsers(c *gin.Context) {
	// 调用服务层获取在线用户列表
	onlineUsers, err := h.onlineService.GetOnlineUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // 错误处理
		return
	}

	// 返回在线用户列表
	c.JSON(http.StatusOK, gin.H{"online_users": onlineUsers})
}

// CheckUserOnline：检查指定用户是否在线
func (h *OnlineHandler) CheckUserOnline(c *gin.Context) {
	// 获取路径参数中的用户ID
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"}) // 如果没有传递用户ID，返回请求错误
		return
	}

	// 调用服务层检查用户是否在线
	isOnline := h.onlineService.IsUserOnline(userID)

	// 返回用户在线状态
	c.JSON(http.StatusOK, gin.H{"is_online": isOnline})
}
