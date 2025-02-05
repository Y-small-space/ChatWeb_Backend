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

type UserService struct {
	userRepo       *repository.UserRepository
	jwtSecret      string
	jwtExpireHours int
}

func NewUserService(userRepo *repository.UserRepository, jwtSecret string, jwtExpireHours int) *UserService {
	return &UserService{
		userRepo:       userRepo,
		jwtSecret:      jwtSecret,
		jwtExpireHours: jwtExpireHours,
	}
}

func (s *UserService) Register(ctx context.Context, user *model.User) error {
	// 检查邮箱是否已存在
	if _, err := s.userRepo.FindByEmail(ctx, user.Email); err == nil {
		return errors.New("email already exists")
	}

	// 检查手机号是否已存在
	if _, err := s.userRepo.FindByPhone(ctx, user.Phone); err == nil {
		return errors.New("phone number already exists")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	return s.userRepo.Create(ctx, user)
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, *model.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, fmt.Errorf("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("invalid email or password")
	}

	// 生成JWT token
	token, err := jwt.GenerateToken(user.ID.Hex(), s.jwtSecret, s.jwtExpireHours)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %v", err)
	}

	return token, user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	return s.userRepo.FindByID(ctx, objID)
}

func (s *UserService) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	return s.userRepo.Update(ctx, objID, updates)
}

func (s *UserService) SearchUser(ctx context.Context, identifier string) (*model.User, error) {
	return s.userRepo.SearchUserByIdentifier(ctx, identifier)
}
