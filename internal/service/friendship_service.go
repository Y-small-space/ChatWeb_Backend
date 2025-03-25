package service

import (
	"chatweb/internal/model"
	"chatweb/internal/repository"
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FriendshipService 管理用户之间的好友关系
type FriendshipService struct {
	// friendshipRepo 用于与数据库交互，管理好友关系数据
	friendshipRepo *repository.FriendshipRepository
	// userRepo 用于与数据库交互，管理用户数据
	userRepo *repository.UserRepository
}

// NewFriendshipService 创建并返回一个新的 FriendshipService 实例
func NewFriendshipService(friendshipRepo *repository.FriendshipRepository, userRepo *repository.UserRepository) *FriendshipService {
	return &FriendshipService{
		friendshipRepo: friendshipRepo,
		userRepo:       userRepo,
	}
}

// SendFriendRequest 发送好友请求
// 输入: userID - 当前用户的ID, friendID - 想要添加为好友的用户ID
// 功能: 1. 检查好友是否存在
//  2. 检查是否尝试添加自己为好友
//  3. 检查是否已经是好友
//  4. 如果条件满足，则在数据库中创建好友关系记录
func (s *FriendshipService) SendFriendRequest(ctx context.Context, userID, friendID string) error {
	// 将用户和朋友的字符串ID转换为 ObjectID
	userObjID, _ := primitive.ObjectIDFromHex(userID)
	friendObjID, _ := primitive.ObjectIDFromHex(friendID)

	// 查询朋友是否存在
	friend, err := s.userRepo.FindByID(ctx, friendObjID)
	if err != nil {
		return errors.New("user not found") // 如果没有找到用户，则返回错误
	}

	// 判断是否尝试添加自己为好友
	if friend.ID.Hex() == userID {
		return errors.New("cannot add yourself as friend")
	}

	// 查询当前用户的好友列表
	friendships, err := s.friendshipRepo.GetFriendsList(ctx, userObjID)
	if err != nil {
		return err // 如果获取好友列表失败，返回错误
	}

	// 检查是否已经是好友
	for _, f := range friendships {
		// 如果已经存在这条好友关系，则返回错误
		if (f.UserID == userObjID && f.FriendID == friend.ID) ||
			(f.UserID == friend.ID && f.FriendID == userObjID) {
			return errors.New("already friends")
		}
	}

	// 创建新的好友关系记录
	friendship := &model.Friendship{
		UserID:   userObjID,
		FriendID: friend.ID,
	}

	// 在数据库中创建这条好友关系记录
	return s.friendshipRepo.Create(ctx, friendship)
}

// GetFriendsList 获取指定用户的好友列表
// 输入: userID - 用户的ID
// 输出: 好友列表，或在出错时返回错误
func (s *FriendshipService) GetFriendsList(ctx context.Context, userID string) ([]*model.User, error) {
	// 将字符串ID转换为 ObjectID
	userObjID, _ := primitive.ObjectIDFromHex(userID)

	// 获取当前用户的所有好友关系
	friendships, err := s.friendshipRepo.GetFriendsList(ctx, userObjID)
	if err != nil {
		return nil, err // 如果获取好友列表失败，返回错误
	}

	// 创建一个切片来存储好友信息
	var friends []*model.User
	for _, f := range friendships {
		// 确定朋友的ID
		var friendID primitive.ObjectID
		if f.UserID == userObjID {
			friendID = f.FriendID
		} else {
			friendID = f.UserID
		}

		// 查询好友的详细信息
		friend, err := s.userRepo.FindByID(ctx, friendID)
		if err != nil {
			continue // 如果找不到好友，则跳过
		}
		friends = append(friends, friend) // 将好友添加到列表中
	}

	// 返回好友列表
	return friends, nil
}

// DeleteFriend 删除好友
func (s *FriendshipService) DeleteFriend(ctx context.Context, friendId string, userId string) error {
	log.Print("delete")
	log.Print("user", friendId)
	// 将用户和好友的字符串 ID 转换为 ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return errors.New("invalid user ID")
	}
	friendObjID, err := primitive.ObjectIDFromHex(friendId)
	if err != nil {
		return errors.New("invalid friend ID")
	}
	// 检查好友关系是否存在
	friendships, err := s.friendshipRepo.GetFriendsList(ctx, userObjID)
	if err != nil {
		return err
	}
	found := false
	for _, f := range friendships {
		if (f.UserID == userObjID && f.FriendID == friendObjID) ||
			(f.UserID == friendObjID && f.FriendID == userObjID) {
			found = true
			break
		}
	}
	if !found {
		return errors.New("friend relationship not found")
	}
	// 从数据库中删除好友关系
	err = s.friendshipRepo.Delete(ctx, userObjID, friendObjID)
	if err != nil {
		return err
	}

	return nil
}
