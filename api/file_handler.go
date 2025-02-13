package api

import (
	"chatweb/internal/service"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	// 服务层，用于处理文件上传、删除等操作
	fileService *service.FileService
}

// NewFileHandler 构造函数，初始化 FileHandler
func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// Upload 处理文件上传请求
func (h *FileHandler) Upload(c *gin.Context) {
	// 获取当前用户的 ID，确保用户已认证
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 检查文件大小，限制为 20MB
	if file.Size > 20*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
		return
	}

	// 检查文件扩展名是否允许
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".mp4":  true,
		".avi":  true,
		".mp3":  true,
		".wav":  true,
		".doc":  true,
		".docx": true,
		".pdf":  true,
	}

	// 如果文件类型不在允许范围内，返回错误
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File type not allowed"})
		return
	}

	// 调用服务层方法处理文件上传
	uploadedFile, err := h.fileService.UploadFile(c.Request.Context(), file, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 上传成功，返回文件信息
	c.JSON(http.StatusOK, gin.H{"file": uploadedFile})
}

// GetUserFiles 获取当前用户上传的文件列表
func (h *FileHandler) GetUserFiles(c *gin.Context) {
	// 获取当前用户的 ID，确保用户已认证
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 获取当前用户上传的所有文件
	files, err := h.fileService.GetUserFiles(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回文件列表
	c.JSON(http.StatusOK, gin.H{"files": files})
}

// Delete 删除指定文件
func (h *FileHandler) Delete(c *gin.Context) {
	// 获取当前用户的 ID，确保用户已认证
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 获取文件 ID，确保参数有效
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File ID is required"})
		return
	}

	// 调用服务层方法删除文件
	if err := h.fileService.DeleteFile(c.Request.Context(), fileID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 删除成功，返回成功消息
	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}
