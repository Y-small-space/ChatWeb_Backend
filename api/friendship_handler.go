package api

import (
	"chatweb/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FriendshipHandler struct {
	friendshipService *service.FriendshipService
}

func NewFriendshipHandler(friendshipService *service.FriendshipService) *FriendshipHandler {
	return &FriendshipHandler{
		friendshipService: friendshipService,
	}
}

func (h *FriendshipHandler) SendRequest(c *gin.Context) {
	var req struct {
		FriendID string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	if err := h.friendshipService.SendFriendRequest(c.Request.Context(), userID, req.FriendID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Friend request sent successfully",
	})
}

func (h *FriendshipHandler) GetFriendsList(c *gin.Context) {
	userID := c.GetString("userID")
	friends, err := h.friendshipService.GetFriendsList(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"friends": friends,
		},
	})
}
