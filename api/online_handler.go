package api

import (
	"chatweb/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OnlineHandler struct {
	onlineService *service.OnlineService
}

func NewOnlineHandler(onlineService *service.OnlineService) *OnlineHandler {
	return &OnlineHandler{
		onlineService: onlineService,
	}
}

func (h *OnlineHandler) GetOnlineUsers(c *gin.Context) {
	onlineUsers, err := h.onlineService.GetOnlineUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"online_users": onlineUsers})
}

func (h *OnlineHandler) CheckUserOnline(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	isOnline := h.onlineService.IsUserOnline(userID)
	c.JSON(http.StatusOK, gin.H{"is_online": isOnline})
}
