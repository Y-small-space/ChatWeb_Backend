package api

import (
	"chatweb/internal/model"   // 引入模型层
	"chatweb/internal/service" // 引入服务层
	"io"
	"net/http" // HTTP 状态码

	"github.com/gin-gonic/gin" // Gin 框架
)

// 允许的图片格式
var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

// UserHandler：处理与用户相关的 API 请求
type UserHandler struct {
	userService *service.UserService // 引入用户服务
}

// NewUserHandler：构造函数，用于初始化 UserHandler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService, // 注入用户服务
	}
}

// RegisterRequest：注册请求体结构体，包含必要的字段和验证规则
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`    // 必须是有效的电子邮件
	Password string `json:"password" binding:"required,min=6"` // 密码要求至少6个字符
	Username string `json:"username" binding:"required"`       // 用户名是必填项
	Phone    string `json:"phone" binding:"required"`          // 手机号是必填项
}

// Register：用户注册接口
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // 如果请求体格式不正确，返回错误
		return
	}

	// 创建新的用户对象
	user := &model.User{
		Email:    req.Email,
		Password: req.Password,
		Username: req.Username,
		Phone:    req.Phone,
	}

	// 调用服务层的注册方法
	if err := h.userService.Register(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // 如果注册失败，返回错误
		return
	}

	// 返回注册成功的响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "User registered successfully",
		"data": map[string]interface{}{
			"user_id":    user.ID.Hex(),
			"username":   user.Username,
			"email":      user.Email,
			"phone":      user.Phone,
			"created_at": user.CreatedAt,
		},
	})
}

// LoginRequest：登录请求体结构体
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"` // 必须是有效的电子邮件
	Password string `json:"password" binding:"required"`    // 密码是必填项
}

// Login：用户登录接口
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // 如果请求体格式不正确，返回错误
		return
	}

	// 调用服务层的登录方法
	token, user, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()}) // 登录失败，返回未授权错误
		return
	}

	// 返回登录成功的响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Login successful",
		"data": map[string]interface{}{
			"token":      token,
			"user_id":    user.ID.Hex(),
			"username":   user.Username,
			"email":      user.Email,
			"phone":      user.Phone,
			"created_at": user.CreatedAt,
			"avatar":     user.Avatar,
		},
	})
}

// GetProfile：获取用户的个人资料
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"}) // 如果用户未授权，返回未授权错误
		return
	}

	// 调用服务层获取用户信息
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // 如果获取失败，返回内部错误
		return
	}

	// 返回用户信息
	c.JSON(http.StatusOK, user)
}

// GetProfile：根据一组id获取这些人的资料
// GetUsersByIDs 批量获取用户信息
func (h *UserHandler) GetUsersByIDs(c *gin.Context) {
	// 解析请求体
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// 调用服务层获取用户信息
	users, failedIDs, err := h.userService.GetUsersByIDs(c.Request.Context(), req.UserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回批量查询结果
	if len(failedIDs) > 0 {
		c.JSON(http.StatusPartialContent, gin.H{
			"message":       "Some users not found",
			"users":         users,
			"failed_ids":    failedIDs,
			"success_count": len(users),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// UpdateProfile：更新用户的个人资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"}) // 如果用户未授权，返回未授权错误
		return
	}

	// 请求体结构体，包含昵称、头像和状态
	var req struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
		Status   string `json:"status"`
	}

	// 绑定请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // 如果请求体格式不正确，返回错误
		return
	}

	// 准备更新的数据
	updates := map[string]interface{}{
		"nickname": req.Nickname,
		"avatar":   req.Avatar,
		"status":   req.Status,
	}

	// 调用服务层更新用户信息
	if err := h.userService.UpdateUser(c.Request.Context(), userID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // 更新失败，返回错误
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// SearchUser：根据查询条件搜索用户
func (h *UserHandler) SearchUser(c *gin.Context) {
	query := c.Query("query") // 获取查询参数
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query is required"}) // 如果查询条件为空，返回错误
		return
	}

	// 调用服务层进行用户搜索
	user, err := h.userService.SearchUser(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "User not found", // 如果没有找到用户，返回 404
		})
		return
	}

	// 返回搜索到的用户
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"user": user,
		},
	})
}

// 处理头像上传
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID := c.PostForm("userId") // 👈 现在可以正常获取

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// 读取文件数据
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// 处理头像上传
	fileURL, err := h.userService.UploadAvatar(c.Request.Context(), userID, header.Filename, fileData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回头像 URL
	c.JSON(http.StatusOK, gin.H{"url": fileURL, "message": "Avatar updated successfully"})
}
