package service

import (
	"chatweb/internal/model"
	"chatweb/internal/repository"
	"chatweb/pkg/jwt"
	"context"
	"errors"
	"fmt"

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
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err // 如果ID格式无效，返回错误
	}
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
