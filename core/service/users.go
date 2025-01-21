package service

import (
	"messanger/core/ports"
	"messanger/models"
	tr "messanger/pkg/error_trace"
)

type UsersService struct {
	repo ports.UsersRepo
}

func NewUsersService(repo ports.UsersRepo) *UsersService {
	return &UsersService{repo: repo}
}

func (s *UsersService) CreateUser(user *models.User) error {
	if err := s.repo.New(user); err != nil {
		return err
	}
	return nil
}

func (s *UsersService) UpdateUser(user *models.User) error {
	if err := s.repo.Update(user); err != nil {
		return tr.Trace(err)
	}
	return nil
}

func (s *UsersService) GetUser(id int) (*models.User, error) {
	user, err := s.repo.GetById(id)
	if err != nil {
		return nil, tr.Trace(err)
	}
	return user, nil
}

func (s *UsersService) DeleteUser(id int) error {
	if err := s.repo.Delete(id); err != nil {
		return tr.Trace(err)
	}
	return nil
}
