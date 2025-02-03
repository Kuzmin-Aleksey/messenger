package messages

import (
	"context"
	"fmt"
	"messanger/domain/models"
	"messanger/domain/ports"
	"messanger/domain/service/auth"
	"messanger/pkg/errors"
	"net/http"
	"time"
)

type UserChats interface {
	GetChatListByUser(ctx context.Context, userId int) ([]int, *errors.Error)
	CheckUserInChat(ctx context.Context, userId int, chatId int) (bool, *errors.Error)
}

type MessagesService struct {
	repo        ports.MessagesRepo
	userChats   UserChats
	connManager *ConnectionsManager
}

func NewMessagesService(repo ports.MessagesRepo, userChecker UserChats) *MessagesService {
	return &MessagesService{
		repo:      repo,
		userChats: userChecker,
	}
}

func (s *MessagesService) CreateMessage(ctx context.Context, m *models.Message) *errors.Error {
	userId := auth.ExtractUser(ctx)
	ok, err := s.userChats.CheckUserInChat(ctx, userId, m.ChatId)
	if err != nil {
		return err.Trace()
	}
	if !ok {
		return errors.New(fmt.Sprintf("user (%d) tried to create a message in the chat (%d)", userId, m.ChatId),
			models.ErrPermissionDenied, http.StatusForbidden)
	}
	m.UserId = userId
	m.Time = time.Now().UTC()
	if err := s.repo.New(ctx, m); err != nil {
		return err.Trace()
	}

	if s.connManager != nil {
		go s.connManager.onCreateMessage(m)
	}
	return nil
}

func (s *MessagesService) UpdateMessage(ctx context.Context, id int, dto *UpdateMessageDTO) *errors.Error {
	if len(dto.Text) == 0 {
		return errors.New1Msg("invalid message", http.StatusBadRequest)
	}
	userId := auth.ExtractUser(ctx)
	ok, err := s.repo.IsUserMessage(ctx, id, userId)
	if err != nil {
		return err.Trace()
	}
	if !ok {
		return errors.New(fmt.Sprintf("user (%d) tried to update a message (%d)", userId, id),
			models.ErrPermissionDenied, http.StatusForbidden)
	}
	if err := s.repo.Update(ctx, id, dto.Text); err != nil {
		return err.Trace()
	}
	m, err := s.repo.GetById(ctx, id)
	if err != nil {
		return err.Trace()
	}
	if s.connManager != nil {
		go s.connManager.onUpdateMessage(m)
	}
	return nil
}

func (s *MessagesService) DeleteMessage(ctx context.Context, id int) *errors.Error {
	if id <= 0 {
		return errors.New1Msg("missing message id", http.StatusBadRequest)
	}
	userId := auth.ExtractUser(ctx)
	m, err := s.repo.GetById(ctx, id)
	if err != nil {
		return err.Trace()
	}
	if m.UserId != userId {
		return errors.New(fmt.Sprintf("user (%d) tried to delete a message (%d)", userId, id),
			models.ErrPermissionDenied, http.StatusForbidden)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err.Trace()
	}
	if s.connManager != nil {
		go s.connManager.onDeleteMessage(id, m.ChatId)
	}
	return nil
}

func (s *MessagesService) GetFromChat(ctx context.Context, dto *GetMessagesDTO) ([]MessagesResponseDTO, *errors.Error) {
	if dto.Count <= 0 {
		return nil, errors.New1Msg("field count is missing", http.StatusBadRequest)
	}
	userId := auth.ExtractUser(ctx)
	ok, err := s.userChats.CheckUserInChat(ctx, userId, dto.ChatId)
	if err != nil {
		return nil, err.Trace()
	}
	if !ok {
		return nil, errors.New(fmt.Sprintf("user (%d) tried to get a messages in the chat (%d)", userId, dto.ChatId),
			models.ErrPermissionDenied, http.StatusForbidden)
	}
	messages, err := s.repo.GetByChat(ctx, dto.ChatId, dto.LastMessageId, dto.Count)
	if err != nil {
		return nil, err.Trace()
	}

	resp := make([]MessagesResponseDTO, len(messages))
	for i := range messages {
		resp[i] = MessagesResponseDTO{
			Id:     messages[i].Id,
			UserId: messages[i].UserId,
			Text:   messages[i].Text,
			Time:   messages[i].Time,
		}
	}

	return resp, nil
}

// GetById for system usage!
func (s *MessagesService) GetById(ctx context.Context, id int) (*models.Message, *errors.Error) {
	m, err := s.repo.GetById(ctx, id)
	if err != nil {
		return nil, err.Trace()
	}
	return m, nil
}

func (s *MessagesService) GetMinMassageIdInChat(ctx context.Context, chatId int) (int, *errors.Error) {
	id, err := s.repo.GetMinMassageIdInChat(ctx, chatId)
	if err != nil {
		return 0, err.Trace()
	}
	return id, nil
}
