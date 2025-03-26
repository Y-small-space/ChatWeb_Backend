package service

import (
	"chatweb/internal/model"
	"chatweb/internal/repository"
	"chatweb/pkg/jwt"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// UserService 提供用户相关的操作服务
type UserService struct {
	userRepo       *repository.UserRepository // 用户存储库，用于与数据库交互
	jwtSecret      string                     // JWT的密钥，用于生成token
	jwtExpireHours int                        // JWT的过期时间，单位小时
}

// NewUserService 创建一个新的 UserService 实例
func NewUserService(userRepo *repository.UserRepository, jwtSecret string, jwtExpireHours int) *UserService {
	return &UserService{
		userRepo:       userRepo,       // 初始化用户存储库
		jwtSecret:      jwtSecret,      // 设置JWT密钥
		jwtExpireHours: jwtExpireHours, // 设置JWT的过期时间
	}
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, user *model.User) error {
	// 检查邮箱是否已存在
	if _, err := s.userRepo.FindByEmail(ctx, user.Email); err == nil {
		return errors.New("email already exists") // 如果邮箱已存在，返回错误
	}

	// 检查手机号是否已存在
	if _, err := s.userRepo.FindByPhone(ctx, user.Phone); err == nil {
		return errors.New("phone number already exists") // 如果手机号已存在，返回错误
	}

	// 加密用户密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err // 密码加密失败，返回错误
	}
	user.Password = string(hashedPassword) // 将加密后的密码存储在用户模型中

	// 将用户数据存入数据库
	return s.userRepo.Create(ctx, user)
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, email, password string) (string, *model.User, error) {
	// 根据邮箱查找用户
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, fmt.Errorf("invalid email or password") // 如果用户不存在，返回错误
	}

	// 校验密码是否正确
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("invalid email or password") // 密码不匹配，返回错误
	}

	// 生成JWT token
	token, err := jwt.GenerateToken(user.ID.Hex(), s.jwtSecret, s.jwtExpireHours)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %v", err) // 如果生成token失败，返回错误
	}

	return token, user, nil // 返回token和用户信息
}

// GetUserByID 根据用户ID获取用户信息
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	log.Print("userId:", userID)
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err // 如果ID格式无效，返回错误
	}
	log.Print("objId", objID)
	return s.userRepo.FindByID(ctx, objID) // 从数据库中获取用户信息
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err // 如果ID格式无效，返回错误
	}
	return s.userRepo.Update(ctx, objID, updates) // 更新用户数据
}

// SearchUser 根据标识符（邮箱或手机号）查找用户
func (s *UserService) SearchUser(ctx context.Context, identifier string) (*model.User, error) {
	return s.userRepo.SearchUserByIdentifier(ctx, identifier) // 根据标识符查找用户
}

// GetUsersByIDs 通过一组 ID 获取多个用户信息
func (s *UserService) GetUsersByIDs(ctx context.Context, userIDs []string) ([]*model.User, []string, error) {
	var objectIDs []primitive.ObjectID
	var failedIDs []string

	// 将字符串 ID 转换为 ObjectID
	for _, id := range userIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			failedIDs = append(failedIDs, id) // 记录转换失败的 ID
			continue
		}
		objectIDs = append(objectIDs, objID)
	}

	// 如果所有 ID 都无效，则直接返回错误
	if len(objectIDs) == 0 {
		return nil, failedIDs, errors.New("no valid user IDs provided")
	}

	// 从数据库中批量查询用户信息
	users, err := s.userRepo.FindByIDs(ctx, objectIDs)
	if err != nil {
		return nil, failedIDs, err
	}

	// 确保所有请求的用户 ID 都匹配返回的数据
	foundIDs := make(map[string]bool)
	for _, user := range users {
		foundIDs[user.ID.Hex()] = true
	}

	// 检查哪些 ID 没有找到
	for _, id := range userIDs {
		if !foundIDs[id] {
			failedIDs = append(failedIDs, id)
		}
	}

	return users, failedIDs, nil
}

// 允许的图片扩展名
var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

// UploadAvatar 处理头像上传
func (s *UserService) UploadAvatar(ctx context.Context, userID string, fileName string, fileData []byte) (string, error) {
	// 检查扩展名
	ext := strings.ToLower(filepath.Ext(fileName))
	if !allowedExtensions[ext] {
		return "", fmt.Errorf("only .jpg, .jpeg, .png files are allowed")
	}

	// 创建目录
	uploadDir := "uploads/avatar"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create upload directory")
	}

	// 生成唯一文件名
	newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, newFileName)

	// 写入文件
	if err := os.WriteFile(savePath, fileData, 0644); err != nil {
		return "", fmt.Errorf("failed to save file")
	}

	// 生成头像 URL
	baseURL := "http://localhost:8080"
	fileURL := fmt.Sprintf("%s/uploads/avatar/%s", baseURL, newFileName)

	// 更新数据库
	if err := s.userRepo.UpdateUserAvatar(ctx, userID, fileURL); err != nil {
		return "", fmt.Errorf("failed to update user avatar: %v", err)
	}

	return fileURL, nil
}
