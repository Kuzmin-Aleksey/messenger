package service

import (
	"messanger/core/ports"
	"messanger/models"
	tr "messanger/pkg/error_trace"
)

type MessagesService struct {
	repo ports.MessagesRepo
}

func NewMessagesService(repo ports.MessagesRepo) *MessagesService {
	return &MessagesService{
		repo: repo,
	}
}

func (s *MessagesService) CreateMessage(m *models.Message) error {
	if err := s.repo.New(m); err != nil {
		return tr.Trace(err)
	}
	return nil
}

func (s *MessagesService) UpdateMessage(m *models.Message) error {
	if err := s.repo.Update(m.Id, m.Text); err != nil {
		return tr.Trace(err)
	}
	return nil
}

func (s *MessagesService) DeleteMessage(id int) error {
	if err := s.repo.Delete(id); err != nil {
		return tr.Trace(err)
	}
	return nil
}

func (s *MessagesService) GetById(id int) (*models.Message, error) {
	message, err := s.repo.GetById(id)
	if err != nil {
		return nil, tr.Trace(err)
	}
	return message, nil
}

func (s *MessagesService) GetFromChat(chatId int, lastId int, count int) ([]models.Message, error) {
	messages, err := s.repo.GetByChat(chatId, lastId, count)
	if err != nil {
		return nil, tr.Trace(err)
	}
	return messages, nil
}
