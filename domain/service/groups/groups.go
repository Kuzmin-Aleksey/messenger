package groups

import (
	"context"
	"fmt"
	"messanger/domain/models"
	"messanger/domain/ports"
	"messanger/domain/service/auth"
	"messanger/pkg/db"
	"messanger/pkg/errors"
	"net/http"
	"slices"
)

type GroupService struct {
	chatsRepo  ports.ChatsRepo
	groupsRepo ports.GroupsRepo
}

func NewGroupService(chatsRepo ports.ChatsRepo, groupsRepo ports.GroupsRepo) *GroupService {
	return &GroupService{
		chatsRepo:  chatsRepo,
		groupsRepo: groupsRepo,
	}
}

func (s *GroupService) NewGroup(ctx context.Context, group *models.Group) (err *errors.Error) {
	if len(group.Name) == 0 {
		return errors.New1Msg("group name is missing", http.StatusBadRequest)
	}
	adminId := auth.ExtractUser(ctx)
	chat := &models.Chat{Type: models.ChatTypeGroup}

	ctx, err = db.WithTx(ctx, s.chatsRepo)
	if err != nil {
		return err.Trace()
	}
	defer db.CommitOnDefer(ctx, &err)

	if err := s.chatsRepo.New(ctx, chat); err != nil {
		return err.Trace()
	}

	group.ChatId = chat.Id
	if err := s.groupsRepo.New(ctx, group); err != nil {
		return err.Trace()
	}
	if err := s.groupsRepo.SetRole(ctx, adminId, group.Id, models.RoleAdmin); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *GroupService) UpdateGroup(ctx context.Context, groupId int, dto *UpdateGroupDTO) *errors.Error {
	if len(dto.Name) == 0 {
		return errors.New1Msg("group name is missing", http.StatusBadRequest)
	}

	actionerId := auth.ExtractUser(ctx)
	role, err := s.groupsRepo.GetRole(ctx, actionerId, groupId)
	if err != nil {
		return err.Trace()
	}
	if role != models.RoleAdmin {
		return errors.New(fmt.Sprintf("user (%d) tried rename group (%d)", actionerId, groupId),
			models.ErrPermissionDenied, http.StatusForbidden)
	}

	if err := s.groupsRepo.Update(ctx, &models.Group{
		Id:   groupId,
		Name: dto.Name,
	}); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *GroupService) AddUserToGroup(ctx context.Context, groupId int, userId int) (err *errors.Error) {
	if userId == 0 {
		return errors.New1Msg("userId is missing", http.StatusBadRequest)
	}
	actionerId := auth.ExtractUser(ctx)
	role, err := s.groupsRepo.GetRole(ctx, actionerId, groupId)
	if err != nil {
		return err.Trace()
	}
	if role != models.RoleAdmin {
		return errors.New(fmt.Sprintf("user (%d) tried add user to group (%d)", actionerId, groupId),
			models.ErrPermissionDenied, http.StatusForbidden)
	}
	group, err := s.groupsRepo.GetById(ctx, groupId)
	if err != nil {
		return err.Trace()
	}

	ctx, err = db.WithTx(ctx, s.chatsRepo)
	if err != nil {
		return err.Trace()
	}
	defer db.CommitOnDefer(ctx, &err)

	if err := s.chatsRepo.AddUserToChat(ctx, group.ChatId, userId); err != nil {
		return err.Trace()
	}
	if err := s.groupsRepo.SetRole(ctx, group.Id, groupId, models.RoleMember); err != nil {

	}
	return nil
}

func (s *GroupService) RemoveUserFromGroup(ctx context.Context, groupId int, userId int) (err *errors.Error) {
	if userId == 0 {
		return errors.New1Msg("userId is missing", http.StatusBadRequest)
	}
	actionerId := auth.ExtractUser(ctx)
	role, err := s.groupsRepo.GetRole(ctx, actionerId, groupId)
	if err != nil {
		return err.Trace()
	}
	userRole, err := s.groupsRepo.GetRole(ctx, userId, groupId)
	if err != nil {
		if err.Code != http.StatusNotFound {
			return err.Trace()
		}
		userRole = models.RoleMember
	}

	group, err := s.groupsRepo.GetById(ctx, groupId)
	if err != nil {
		return err.Trace()
	}
	if role != models.RoleAdmin && actionerId != userId {
		return errors.New(fmt.Sprintf("user (%d) tried remove user from group (%d)", actionerId, groupId),
			models.ErrPermissionDenied, http.StatusForbidden)
	}
	if role == userRole && role == models.RoleAdmin {
		return errors.New(fmt.Sprintf("admin (%d) tried remove admin (%d) from group (%d)", actionerId, userId, groupId),
			models.ErrPermissionDenied, http.StatusForbidden)
	}

	ctx, err = db.WithTx(ctx, s.chatsRepo)
	if err != nil {
		return err.Trace()
	}
	defer db.CommitOnDefer(ctx, &err)

	if err := s.chatsRepo.RemoveUserFromChat(ctx, group.ChatId, userId); err != nil {
		return err.Trace()
	}

	count, err := s.chatsRepo.CountUsersInChat(ctx, group.ChatId)
	if err != nil {
		return err.Trace()
	}
	if count == 0 {
		if err := s.chatsRepo.Delete(ctx, group.ChatId); err != nil {
			return err.Trace()
		}
	}
	return nil
}

func (s *GroupService) SetUsersRole(ctx context.Context, groupId int, userId int, role string) *errors.Error {
	if !models.ValidateRole(role) {
		return errors.New1Msg("invalid role: "+role, http.StatusBadRequest)
	}
	actionerId := auth.ExtractUser(ctx)
	actionerRole, err := s.groupsRepo.GetRole(ctx, actionerId, groupId)
	if err != nil {
		return err.Trace()
	}
	if actionerRole != models.RoleAdmin {
		return errors.New(fmt.Sprintf("user (%d) tried set role in group (%d)", actionerId, groupId),
			models.ErrPermissionDenied, http.StatusForbidden)
	}

	if err := s.groupsRepo.SetRole(ctx, userId, groupId, role); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *GroupService) GetUsersByGroup(ctx context.Context, groupId int) ([]GetUsersDTO, *errors.Error) {
	if groupId == 0 {
		return nil, errors.New1Msg("groupId is missing", http.StatusBadRequest)
	}
	users, err := s.groupsRepo.GetUsersByGroup(ctx, groupId)
	if err != nil {
		return nil, err.Trace()
	}
	actionerId := auth.ExtractUser(ctx)
	if slices.Index(users, actionerId) == -1 {
		return nil, errors.New(fmt.Sprintf("user (%d) tried get users from group (%d)", actionerId, groupId),
			models.ErrPermissionDenied, http.StatusForbidden)
	}

	var resp []GetUsersDTO
	for _, user := range users {
		role, err := s.groupsRepo.GetRole(ctx, user, groupId)
		if err != nil {
			if err.Code == http.StatusNotFound {
				role = models.RoleMember
			} else {
				return nil, err.Trace()
			}
		}
		resp = append(resp, GetUsersDTO{
			UserId: user,
			Role:   role,
		})
	}

	return resp, nil
}

func (s *GroupService) RemoveGroup(ctx context.Context, groupId int) (err *errors.Error) {
	actionerId := auth.ExtractUser(ctx)
	role, err := s.groupsRepo.GetRole(ctx, actionerId, groupId)
	if err != nil {
		return err.Trace()
	}
	if role != models.RoleAdmin {
		return errors.New(fmt.Sprintf("user (%d) tried remove group (%d)", actionerId, groupId),
			models.ErrPermissionDenied, http.StatusForbidden)
	}

	group, err := s.groupsRepo.GetById(ctx, groupId)
	if err != nil {
		return err.Trace()
	}

	ctx, err = db.WithTx(ctx, s.chatsRepo)
	if err != nil {
		return err.Trace()
	}
	defer db.CommitOnDefer(ctx, &err)

	if err := s.groupsRepo.Delete(ctx, group.Id); err != nil {
		return err.Trace()
	}
	if err := s.chatsRepo.Delete(ctx, group.ChatId); err != nil {
		return err.Trace()
	}
	return nil
}
