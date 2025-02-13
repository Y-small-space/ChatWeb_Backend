package api

import (
	"chatweb/internal/model"
	"chatweb/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GroupHandler 处理与群组相关的请求，例如创建群组、获取群组列表、加入/离开群组等
type GroupHandler struct {
	// 服务层，用于处理与群组相关的业务逻辑
	groupService *service.GroupService
}

// NewGroupHandler 构造函数，初始化 GroupHandler
func NewGroupHandler(groupService *service.GroupService) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
	}
}

// CreateGroupRequest 用于请求创建群组的参数结构体
type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"` // 群组名称
	Description string `json:"description"`             // 群组描述
}

// Create 创建群组
func (h *GroupHandler) Create(c *gin.Context) {
	// 获取当前用户的 ID，确保用户已认证
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 绑定请求数据
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 将用户 ID 转换为 ObjectID 类型
	creatorID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// 创建群组对象
	group := &model.Group{
		Name:        req.Name,
		Description: req.Description,
		CreatorID:   creatorID,
		Members:     []primitive.ObjectID{creatorID}, // 默认创建者为群组成员
	}

	// 调用服务层创建群组
	if err := h.groupService.CreateGroup(c.Request.Context(), group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回创建成功的群组信息
	c.JSON(http.StatusOK, gin.H{"group": group})
}

// List 获取当前用户所有群组的列表
func (h *GroupHandler) List(c *gin.Context) {
	// 获取当前用户的 ID，确保用户已认证
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 获取用户所属的群组列表
	groups, err := h.groupService.GetUserGroups(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回群组列表
	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

// Get 获取指定群组的详细信息，包括群组成员
func (h *GroupHandler) Get(c *gin.Context) {
	// 从 URL 参数获取群组 ID
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID is required"})
		return
	}

	// 将群组 ID 转换为 ObjectID 类型
	groupObjID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	// 获取群组信息
	group, err := h.groupService.GetGroupByID(c.Request.Context(), groupObjID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取群组成员信息
	members, err := h.groupService.GetGroupMembers(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回群组信息和成员列表
	c.JSON(http.StatusOK, gin.H{
		"group":   group,
		"members": members,
	})
}

// Join 处理用户加入指定群组的请求
func (h *GroupHandler) Join(c *gin.Context) {
	// 获取当前用户的 ID，确保用户已认证
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 从 URL 参数获取群组 ID
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID is required"})
		return
	}

	// 调用服务层方法加入群组
	if err := h.groupService.JoinGroup(c.Request.Context(), groupID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 返回成功加入群组的消息
	c.JSON(http.StatusOK, gin.H{"message": "Successfully joined the group"})
}

// Leave 处理用户离开指定群组的请求
func (h *GroupHandler) Leave(c *gin.Context) {
	// 获取当前用户的 ID，确保用户已认证
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 从 URL 参数获取群组 ID
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID is required"})
		return
	}

	// 调用服务层方法离开群组
	if err := h.groupService.LeaveGroup(c.Request.Context(), groupID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 返回成功离开群组的消息
	c.JSON(http.StatusOK, gin.H{"message": "Successfully left the group"})
}
