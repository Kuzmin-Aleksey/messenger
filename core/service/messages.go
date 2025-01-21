package service

import (
	"messanger/core/ports"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
)

type MessagesService struct {
	repo ports.MessagesRepo
}

func NewMessagesService(repo ports.MessagesRepo) *MessagesService {
	return &MessagesService{
		repo: repo,
	}
}

func (s *MessagesService) CreateMessage(m *domain.Message) *errors.Error {
	if err := s.repo.New(m); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *MessagesService) UpdateMessage(m *domain.Message) *errors.Error {
	if m.Id == 0 || len(m.Text) == 0 {
		return errors.New1Msg("invalid message", http.StatusBadRequest)
	}
	if err := s.repo.Update(m.Id, m.Text); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *MessagesService) DeleteMessage(id int) *errors.Error {
	if err := s.repo.Delete(id); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *MessagesService) GetById(id int) (*domain.Message, *errors.Error) {
	message, err := s.repo.GetById(id)
	if err != nil {
		return nil, err.Trace()
	}
	return message, nil
}

func (s *MessagesService) GetFromChat(chatId int, lastId int, count int) ([]domain.Message, *errors.Error) {
	messages, err := s.repo.GetByChat(chatId, lastId, count)
	if err != nil {
		return nil, err.Trace()
	}
	return messages, nil
}
