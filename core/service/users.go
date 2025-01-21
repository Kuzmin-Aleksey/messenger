package service

import (
	"messanger/core/ports"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
)

type UsersService struct {
	repo ports.UsersRepo
}

func NewUsersService(repo ports.UsersRepo) *UsersService {
	return &UsersService{repo: repo}
}

func (s *UsersService) CreateUser(user *domain.UserInfo) *errors.Error {
	if len(user.Name) == 0 || len(user.Email) == 0 || len(user.Password) == 0 {
		return errors.New1Msg("invalid user info", http.StatusBadRequest)
	}
	if err := s.repo.New(user); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *UsersService) UpdateUser(user *domain.UserInfo) *errors.Error {
	if user.Id == 0 {
		return errors.New1Msg("user id is required", http.StatusBadRequest)
	}
	if err := s.repo.Update(user); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *UsersService) GetUser(id int) (*domain.User, *errors.Error) {
	if id == 0 {
		return nil, errors.New1Msg("user id is required", http.StatusBadRequest)
	}
	user, err := s.repo.GetById(id)
	if err != nil {
		return nil, err.Trace()
	}
	return user, nil
}

func (s *UsersService) DeleteUser(id int) *errors.Error {
	if id == 0 {
		return errors.New1Msg("user id is required", http.StatusBadRequest)
	}
	if err := s.repo.Delete(id); err != nil {
		return err.Trace()
	}
	return nil
}
