package service

import (
	"messanger/core/ports"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
)

type ChatService struct {
	repo ports.ChatsRepo
}

func NewChatService(repo ports.ChatsRepo) *ChatService {
	return &ChatService{repo: repo}
}

func (s *ChatService) NewChat(chat *domain.Chat) *errors.Error {
	if len(chat.Users) == 0 {
		return errors.New1Msg("no users in chat", http.StatusBadRequest)
	}
	if err := s.repo.New(chat); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *ChatService) UpdateChar(chat *domain.Chat) *errors.Error {
	if err := s.repo.Update(chat); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *ChatService) CheckUserInChat(chatId int, userId int) (bool, *errors.Error) {
	exist, err := s.repo.CheckUserInChat(chatId, userId)
	if err != nil {
		return exist, err.Trace()
	}
	return exist, nil
}

func (s *ChatService) GetChat(chatId int) (*domain.Chat, *errors.Error) {
	chat, err := s.repo.GetById(chatId)
	if err != nil {
		return nil, err.Trace()
	}
	return chat, nil
}

func (s *ChatService) Delete(chatId int) *errors.Error {
	if err := s.repo.Delete(chatId); err != nil {
		return err.Trace()
	}
	return nil
}
