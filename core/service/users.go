package service

import (
	"context"
	"fmt"
	"messanger/core/ports"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
	"time"
)

type ConfirmEmailService interface {
	SetOnConfirm(f func(userId int) *errors.Error)
	ToConfirming(userId int, email string) *errors.Error
}

type UsersService struct {
	repo                ports.UsersRepo
	confirmEmailService ConfirmEmailService
	DeleteChat          chan int
}

func NewUsersService(repo ports.UsersRepo, confirmEmailService ConfirmEmailService) *UsersService {
	s := &UsersService{
		repo:                repo,
		confirmEmailService: confirmEmailService,
		DeleteChat:          make(chan int, 10),
	}
	confirmEmailService.SetOnConfirm(s.OnConfirm)
	return s
}

func (s *UsersService) CreateUser(user *domain.User) *errors.Error {
	if len(user.Name) == 0 || len(user.Email) == 0 || len(user.Password) == 0 {
		return errors.New1Msg("invalid user info", http.StatusBadRequest)
	}
	if _, err := s.repo.GetByEmail(user.Email); err != nil {
		if err.Code != http.StatusNotFound {
			return err
		}
	} else {
		return errors.New1Msg("user already exists", http.StatusBadRequest)
	}

	if err := s.repo.New(user); err != nil {
		return err.Trace()
	}
	if err := s.confirmEmailService.ToConfirming(user.Id, user.Email); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *UsersService) OnConfirm(userId int) *errors.Error {
	if err := s.repo.SetConfirm(userId, true); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *UsersService) UpdateUser(ctx context.Context, user *domain.User, lastPass string) *errors.Error {
	user.Id = getUserId(ctx)
	if len(user.Email) != 0 {
		if err := s.repo.SetConfirm(user.Id, false); err != nil {
			return err.Trace()
		}
		if err := s.confirmEmailService.ToConfirming(user.Id, user.Email); err != nil {
			return err.Trace()
		}
	}
	if len(user.Password) != 0 {
		if len(lastPass) == 0 {
			return errors.New1Msg("missing last password", http.StatusBadRequest)
		}
		if ok, err := s.repo.CheckPass(user.Id, lastPass); !ok {
			if err != nil {
				return err.Trace()
			}
			return errors.New1Msg("invalid password", http.StatusUnauthorized)
		}
	}
	if err := s.repo.Update(user); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *UsersService) GetUserInfo(ctx context.Context) (*domain.User, *errors.Error) {
	user, err := s.repo.GetById(getUserId(ctx))
	if err != nil {
		return nil, err.Trace()
	}
	return user, nil
}

func (s *UsersService) DeleteUser(ctx context.Context) *errors.Error {
	id := getUserId(ctx)
	if err := s.repo.Delete(id); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *UsersService) GetUsersByChat(ctx context.Context, chatId int) ([]domain.User, *errors.Error) {
	actionerId := getUserId(ctx)
	ok, err := s.repo.CheckUserInChat(actionerId, chatId)
	if err != nil {
		return nil, err.Trace()
	}
	if !ok {
		return nil, errors.New(fmt.Sprintf("user (%d) tried get users from chat (%d)", actionerId, chatId),
			domain.ErrPermissionDenied, http.StatusForbidden)
	}
	users, err := s.repo.GetUsersByChat(chatId)
	if err != nil {
		return nil, err.Trace()
	}
	return users, nil
}

func (s *UsersService) AddUserToChat(ctx context.Context, userId int, chatId int, role string) *errors.Error {
	if chatId <= 0 || userId <= 0 {
		return errors.New1Msg("missing user id or chat id", http.StatusBadRequest)
	}
	actionerId := getUserId(ctx)
	actionerRole, err := s.repo.GetRole(actionerId, chatId)
	if err != nil {
		return err.Trace()
	}
	if actionerRole != "admin" {
		return errors.New(fmt.Sprintf("user (%d) tried add user to chat (%d)", actionerId, chatId),
			domain.ErrPermissionDenied, http.StatusForbidden)
	}
	if err := s.repo.AddUserToChat(userId, chatId, role); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *UsersService) SetRole(ctx context.Context, userId int, chatId int, role string) *errors.Error {
	actionerId := getUserId(ctx)
	actionerRole, err := s.repo.GetRole(actionerId, chatId)
	if err != nil {
		return err.Trace()
	}
	if actionerRole != "admin" {
		return errors.New(fmt.Sprintf("user (%d) tried set role (%s) to user (%d)", actionerId, role, userId),
			domain.ErrPermissionDenied, http.StatusForbidden)
	}
	if err := s.repo.SetRole(userId, chatId, role); err != nil {
		return err.Trace()
	}
	return nil
}

// DeleteUserFromChat if last user is deleted, the chat will be deleted.
func (s *UsersService) DeleteUserFromChat(ctx context.Context, userId int, chatId int) *errors.Error {
	if chatId <= 0 || userId <= 0 {
		return errors.New1Msg("missing user id or chat id", http.StatusBadRequest)
	}
	actionerId := getUserId(ctx)
	role, err := s.repo.GetRole(actionerId, chatId)
	if err != nil {
		return err.Trace()
	}
	actioner, err := s.repo.GetById(actionerId)
	if err != nil {
		return err.Trace()
	}
	if actioner.Id != userId && role != "admin" {
		return errors.New(fmt.Sprintf("user (%d) tried to delete a user (%d) in chat (%d)", actionerId, userId, chatId),
			domain.ErrPermissionDenied, http.StatusForbidden)
	}
	if role == "admin" && actioner.Id == userId {
		return errors.New(fmt.Sprintf("user (%d - %s) tried to delete self", actionerId, role),
			"admin can't delete self", http.StatusConflict)
	}
	if err := s.repo.DeleteUserFromChat(userId, chatId); err != nil {
		return err.Trace()
	}

	count, err := s.repo.CountOfUsersInChat(chatId)
	if err != nil {
		return err.Trace()
	}
	if count == 0 {
		timeout := time.Tick(time.Millisecond * 500)
		select {
		case s.DeleteChat <- chatId:
		case <-timeout:
		}
	}
	return nil
}
