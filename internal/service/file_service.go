package service

import (
	"chatweb/internal/model"
	"chatweb/internal/repository"
	"chatweb/pkg/storage"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileService struct {
	fileRepo    *repository.FileRepository
	minioClient *storage.MinioClient
}

func NewFileService(fileRepo *repository.FileRepository, minioClient *storage.MinioClient) *FileService {
	return &FileService{
		fileRepo:    fileRepo,
		minioClient: minioClient,
	}
}

func (s *FileService) UploadFile(ctx context.Context, file *multipart.FileHeader, userID string) (*model.File, error) {
	// 生成唯一的文件名
	ext := filepath.Ext(file.Filename)
	objectName := fmt.Sprintf("%s-%d%s", strings.TrimSuffix(file.Filename, ext), time.Now().Unix(), ext)

	// 上传到MinIO
	url, err := s.minioClient.UploadFile(ctx, file, objectName)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %v", err)
	}

	// 确定文件类型
	fileType := s.determineFileType(ext)

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// 创建文件记录
	fileRecord := &model.File{
		Name:   file.Filename,
		Type:   fileType,
		Size:   file.Size,
		URL:    url,
		UserID: userObjID,
	}

	if err := s.fileRepo.Create(ctx, fileRecord); err != nil {
		// 如果数据库创建失败，删除已上传的文件
		_ = s.minioClient.DeleteFile(ctx, objectName)
		return nil, fmt.Errorf("failed to create file record: %v", err)
	}

	return fileRecord, nil
}

func (s *FileService) GetUserFiles(ctx context.Context, userID string) ([]*model.File, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	return s.fileRepo.GetByUserID(ctx, userObjID)
}

func (s *FileService) DeleteFile(ctx context.Context, fileID string, userID string) error {
	fileObjID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return err
	}

	// 获取文件信息
	file, err := s.fileRepo.GetByID(ctx, fileObjID)
	if err != nil {
		return err
	}

	// 验证文件所有者
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	if file.UserID != userObjID {
		return fmt.Errorf("unauthorized to delete this file")
	}

	// 从MinIO删除文件
	objectName := strings.TrimPrefix(file.URL, "/"+s.minioClient.GetBucketName()+"/")
	if err := s.minioClient.DeleteFile(ctx, objectName); err != nil {
		return fmt.Errorf("failed to delete file from storage: %v", err)
	}

	// 从数据库删除记录
	return s.fileRepo.Delete(ctx, fileObjID)
}

func (s *FileService) determineFileType(ext string) model.FileType {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return model.ImageFile
	case ".mp4", ".avi", ".mov", ".wmv":
		return model.VideoFile
	case ".mp3", ".wav", ".ogg":
		return model.AudioFile
	default:
		return model.DocFile
	}
}

func (s *FileService) Upload(ctx context.Context, file *model.File, reader io.Reader) error {
	if s.minioClient == nil {
		return fmt.Errorf("file upload is not available: MinIO client is not initialized")
	}
	// ... 其他代码
	return nil
}

func (s *FileService) Download(ctx context.Context, fileID string) (io.Reader, error) {
	if s.minioClient == nil {
		return nil, fmt.Errorf("file download is not available: MinIO client is not initialized")
	}
	// ... 其他代码
	return nil, nil
}
