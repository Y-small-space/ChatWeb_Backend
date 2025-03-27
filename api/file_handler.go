package api

import (
	"chatweb/internal/service"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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

func (h *FileHandler) UploadFile(c *gin.Context) {
	log.Print("uploading...")

	// 获取文件
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed"})
		return
	}
	defer file.Close()

	// 获取文件类型和大小
	fileType := fileHeader.Header.Get("Content-Type")
	fileSize := fileHeader.Size

	log.Print("fileType: ", fileType)
	log.Print("fileSize: ", fileSize)

	// 生成唯一文件名
	filename := fmt.Sprintf("%d-%s", time.Now().Unix(), fileHeader.Filename)
	log.Print("filename: ", filename)

	// 设置保存文件的路径
	var filePath string
	if fileType == "image/jpeg" || fileType == "image/png" {
		if fileSize < 2*1024*1024 {
			filePath = "./uploads/avator/" + filename
		} else {
			filePath = "./uploads/files/" + filename
		}
	} else if fileType == "audio/mpeg" || fileType == "audio/wav" {
		if fileSize < 60*1024*1024 {
			filePath = "./uploads/audio/" + filename
		} else {
			filePath = "./uploads/files/" + filename
		}
	} else if fileType == "video/mp4" || fileType == "video/webm" {
		if fileSize < 60*1024*1024 {
			filePath = "./uploads/video/" + filename
		} else {
			filePath = "./uploads/files/" + filename
		}
	} else {
		filePath = "./uploads/files/" + filename
	}

	// 确保目录存在
	err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		log.Print("Failed to create directory: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		log.Print("Failed to create file: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	// 重新设置文件指针，确保文件流未被提前读取完
	file.Seek(0, io.SeekStart)

	// 将文件内容从上传的 file 写入到目标文件
	_, err = io.Copy(dst, file)
	if err != nil {
		log.Print("Failed to write file: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
		return
	}

	// 生成文件的访问URL
	fileURL := strings.TrimPrefix(filePath, "./uploads") // 去掉本地路径前缀，形成 URL 路径
	fileURL = "http://localhost:8080/uploads" + fileURL

	log.Print("File uploaded successfully: ", fileURL)

	// 返回文件的URL
	c.JSON(http.StatusOK, gin.H{"url": fileURL})
}
