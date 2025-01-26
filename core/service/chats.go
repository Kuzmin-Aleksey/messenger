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

type ChatService struct {
	repo         ports.ChatsRepo
	OnDeleteChat chan int // to delete messages from chat
}

func NewChatService(repo ports.ChatsRepo) *ChatService {
	return &ChatService{
		repo:         repo,
		OnDeleteChat: make(chan int, 10),
	}
}

func (s *ChatService) NewChat(ctx context.Context, chat *domain.Chat) *errors.Error {
	if len(chat.Name) == 0 {
		return errors.New1Msg("chat name is missing", http.StatusBadRequest)
	}
	if err := s.repo.New(chat, getUserId(ctx)); err != nil {
		return err.Trace()
	}

	return nil
}

func (s *ChatService) UpdateChat(ctx context.Context, chat *domain.Chat) *errors.Error {
	if chat.Id <= 0 {
		return errors.New1Msg("chat id is missing", http.StatusBadRequest)
	}
	actionerId := getUserId(ctx)
	role, err := s.repo.GetUserRole(actionerId, chat.Id)
	if err != nil {
		return err.Trace()
	}
	if role != "admin" {
		return errors.New(fmt.Sprintf("user (%d) tried to update chat (%d)", actionerId, chat.Id),
			domain.ErrPermissionDenied, http.StatusBadRequest)
	}
	if err := s.repo.Update(chat); err != nil {
		return err.Trace()
	}
	return nil
}

func (s *ChatService) GetUserChats(ctx context.Context) ([]domain.Chat, *errors.Error) {
	chats, err := s.repo.GetChatsByUser(getUserId(ctx))
	if err != nil {
		return nil, err.Trace()
	}
	return chats, nil
}

// Delete write chat id to chan OnDeleteChat
func (s *ChatService) Delete(ctx context.Context, id int) *errors.Error {
	if ctx.Value("IsSystemCall") != nil {
		if err := s.repo.Delete(id); err != nil {
			return err.Trace()
		}
		return nil
	}

	if id <= 0 {
		return errors.New1Msg("chat id is missing", http.StatusBadRequest)
	}
	actionerId := getUserId(ctx)
	role, err := s.repo.GetUserRole(actionerId, id)
	if err != nil {
		return err.Trace()
	}
	if role != "admin" {
		return errors.New(fmt.Sprintf("user (%d) tried to delete chat (%d)", actionerId, id),
			domain.ErrPermissionDenied, http.StatusForbidden)
	}
	if err := s.repo.Delete(id); err != nil {
		return err.Trace()
	}

	timeout := time.Tick(time.Millisecond * 500)
	select {
	case s.OnDeleteChat <- id:
	case <-timeout:
	}
	return nil
}
