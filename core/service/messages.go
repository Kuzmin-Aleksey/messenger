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

type UserChecker interface {
	CheckUserInChat(userId int, chatId int) (bool, *errors.Error)
}

type MessagesService struct {
	repo        ports.MessagesRepo
	userChecker UserChecker
}

func NewMessagesService(repo ports.MessagesRepo, userChecker UserChecker) *MessagesService {
	return &MessagesService{
		repo:        repo,
		userChecker: userChecker,
	}
}

func (s *MessagesService) CreateMessage(ctx context.Context, m *domain.Message) *errors.Error {
	userId := getUserId(ctx)
	ok, err := s.userChecker.CheckUserInChat(userId, m.ChatId)
	if err != nil {
		return err.Trace()
	}
	if !ok {
		return errors.New(fmt.Sprintf("user (%d) tried to create a message in the chat (%d)", userId, m.ChatId),
			domain.ErrPermissionDenied, http.StatusForbidden)
	}
	m.UserId = userId
	m.Time = time.Now()
	if err := s.repo.New(m); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *MessagesService) UpdateMessage(ctx context.Context, m *domain.Message) *errors.Error {
	if m.Id <= 0 || len(m.Text) == 0 {
		return errors.New1Msg("invalid message", http.StatusBadRequest)
	}
	userId := getUserId(ctx)
	messageNow, err := s.repo.GetById(m.Id)
	if err != nil {
		return err.Trace()
	}
	if messageNow.UserId != userId {
		return errors.New(fmt.Sprintf("user (%d) tried to update a message (%d) with user id (%d)", userId, m.Id, m.UserId),
			domain.ErrPermissionDenied, http.StatusForbidden)
	}
	m.UserId = userId
	if err := s.repo.Update(m.Id, m.Text); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *MessagesService) DeleteMessage(ctx context.Context, id int) (*domain.Message, *errors.Error) {
	if id <= 0 {
		return nil, errors.New1Msg("missing message id", http.StatusBadRequest)
	}
	userId := getUserId(ctx)
	m, err := s.repo.GetById(id)
	if err != nil {
		return nil, err.Trace()
	}
	if m.UserId != userId {
		return nil, errors.New(fmt.Sprintf("user (%d) tried to delete a message (%d) with user id (%d)", userId, m.Id, m.UserId),
			domain.ErrPermissionDenied, http.StatusForbidden)
	}
	if err := s.repo.Delete(id); err != nil {
		return nil, err.Trace()
	}
	return m, nil
}

func (s *MessagesService) GetFromChat(ctx context.Context, chatId int, lastId int, count int) ([]domain.Message, *errors.Error) {
	if count <= 0 {
		return nil, errors.New1Msg("field count is missing", http.StatusBadRequest)
	}
	userId := getUserId(ctx)
	ok, err := s.userChecker.CheckUserInChat(userId, chatId)
	if err != nil {
		return nil, err.Trace()
	}
	if !ok {
		return nil, errors.New(fmt.Sprintf("user (%d) tried to get a messages in the chat (%d)", userId, chatId),
			domain.ErrPermissionDenied, http.StatusForbidden)
	}
	messages, err := s.repo.GetByChat(chatId, lastId, count)
	if err != nil {
		return nil, err.Trace()
	}
	return messages, nil
}

func (s *MessagesService) GetMinMassageIdInChat(chatId int) (int, *errors.Error) {
	id, err := s.repo.GetMinMassageIdInChat(chatId)
	if err != nil {
		return 0, err.Trace()
	}
	return id, nil
}

func (s *MessagesService) OnDeleteChat(chatId int) *errors.Error {
	if err := s.repo.DeleteByChat(chatId); err != nil {
		return err.Trace()
	}
	return nil
}
