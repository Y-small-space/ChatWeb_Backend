package service

import (
	"chatweb/internal/model"
	"chatweb/internal/repository"
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupService struct {
	groupRepo *repository.GroupRepository
}

func NewGroupService(groupRepo *repository.GroupRepository) *GroupService {
	return &GroupService{
		groupRepo: groupRepo,
	}
}

func (s *GroupService) CreateGroup(ctx context.Context, group *model.Group) error {
	// 创建群组
	if err := s.groupRepo.Create(ctx, group); err != nil {
		return err
	}

	// 添加创建者为群组成员
	member := &model.GroupMember{
		GroupID: group.ID,
		UserID:  group.CreatorID,
		Role:    "admin",
	}

	return s.groupRepo.AddMember(ctx, member)
}

func (s *GroupService) JoinGroup(ctx context.Context, groupID string, userID string) error {
	groupObjID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return err
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// 检查群组是否存在
	group, err := s.groupRepo.GetGroupByID(ctx, groupObjID)
	if err != nil {
		return err
	}

	// 检查用户是否已经是群组成员
	members, err := s.groupRepo.GetGroupMembers(ctx, groupObjID)
	if err != nil {
		return err
	}

	for _, member := range members {
		if member.UserID == userObjID {
			return errors.New("user is already a member of this group")
		}
	}

	// 添加新成员
	member := &model.GroupMember{
		GroupID: group.ID,
		UserID:  userObjID,
		Role:    "member",
	}

	return s.groupRepo.AddMember(ctx, member)
}

func (s *GroupService) LeaveGroup(ctx context.Context, groupID string, userID string) error {
	groupObjID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return err
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	return s.groupRepo.RemoveMember(ctx, groupObjID, userObjID)
}

func (s *GroupService) GetUserGroups(ctx context.Context, userID string) ([]*model.Group, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	return s.groupRepo.GetGroupsByUserID(ctx, userObjID)
}

func (s *GroupService) GetGroupMembers(ctx context.Context, groupID string) ([]*model.GroupMember, error) {
	groupObjID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, err
	}

	return s.groupRepo.GetGroupMembers(ctx, groupObjID)
}

func (s *GroupService) GetGroupByID(ctx context.Context, groupID primitive.ObjectID) (*model.Group, error) {
	return s.groupRepo.GetGroupByID(ctx, groupID)
}

func (s *GroupService) GetGroupMemberIDs(ctx context.Context, groupID string) ([]string, error) {
	groupObjID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID: %v", err)
	}

	members, err := s.groupRepo.GetGroupMembers(ctx, groupObjID)
	if err != nil {
		return nil, err
	}

	memberIDs := make([]string, len(members))
	for i, member := range members {
		memberIDs[i] = member.UserID.Hex()
	}

	return memberIDs, nil
}
