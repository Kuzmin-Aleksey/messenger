package service

import (
	"messanger/core/ports"
	"messanger/models"
	tr "messanger/pkg/error_trace"
)

type ChatService struct {
	repo ports.ChatsRepo
}

func NewChatService(repo ports.ChatsRepo) *ChatService {
	return &ChatService{repo: repo}
}

func (s *ChatService) NewChat(chat *models.Chat) error {
	if len(chat.Users) == 0 {
		return tr.Trace("no users in chat")
	}
	if err := s.repo.New(chat); err != nil {
		return tr.Trace(err)
	}
	return nil
}

func (s *ChatService) UpdateChar(chat *models.Chat) error {
	if err := s.repo.Update(chat); err != nil {
		return tr.Trace(err)
	}
	return nil
}

func (s *ChatService) CheckUserInChat(chatId int, userId int) (bool, error) {
	exist, err := s.repo.CheckUserInChat(chatId, userId)
	if err != nil {
		return exist, tr.Trace(err)
	}
	return exist, nil
}

func (s *ChatService) GetChat(chatId int) (*models.Chat, error) {
	chat, err := s.repo.GetById(chatId)
	if err != nil {
		return nil, tr.Trace(err)
	}
	return chat, nil
}

func (s *ChatService) Delete(chatId int) error {
	if err := s.repo.Delete(chatId); err != nil {
		return tr.Trace(err)
	}
	return nil
}
