package service

import (
	"chatweb/internal/model"
	"chatweb/internal/repository"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FriendshipService struct {
	friendshipRepo *repository.FriendshipRepository
	userRepo       *repository.UserRepository
}

func NewFriendshipService(friendshipRepo *repository.FriendshipRepository, userRepo *repository.UserRepository) *FriendshipService {
	return &FriendshipService{
		friendshipRepo: friendshipRepo,
		userRepo:       userRepo,
	}
}

func (s *FriendshipService) SendFriendRequest(ctx context.Context, userID, friendID string) error {
	userObjID, _ := primitive.ObjectIDFromHex(userID)
	friendObjID, _ := primitive.ObjectIDFromHex(friendID)

	friend, err := s.userRepo.FindByID(ctx, friendObjID)
	if err != nil {
		return errors.New("user not found")
	}

	if friend.ID.Hex() == userID {
		return errors.New("cannot add yourself as friend")
	}

	friendships, err := s.friendshipRepo.GetFriendsList(ctx, userObjID)
	if err != nil {
		return err
	}
	for _, f := range friendships {
		if (f.UserID == userObjID && f.FriendID == friend.ID) ||
			(f.UserID == friend.ID && f.FriendID == userObjID) {
			return errors.New("already friends")
		}
	}

	friendship := &model.Friendship{
		UserID:   userObjID,
		FriendID: friend.ID,
	}

	return s.friendshipRepo.Create(ctx, friendship)
}

func (s *FriendshipService) GetFriendsList(ctx context.Context, userID string) ([]*model.User, error) {
	userObjID, _ := primitive.ObjectIDFromHex(userID)

	friendships, err := s.friendshipRepo.GetFriendsList(ctx, userObjID)
	if err != nil {
		return nil, err
	}

	var friends []*model.User
	for _, f := range friendships {
		var friendID primitive.ObjectID
		if f.UserID == userObjID {
			friendID = f.FriendID
		} else {
			friendID = f.UserID
		}

		friend, err := s.userRepo.FindByID(ctx, friendID)
		if err != nil {
			continue
		}
		friends = append(friends, friend)
	}

	return friends, nil
}
