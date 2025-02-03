package messages

import (
	"context"
	"fmt"
	"messanger/domain/models"
	"messanger/domain/ports"
	"messanger/domain/service/auth"
	"messanger/pkg/db"
	"messanger/pkg/errors"
	"net/http"
	"time"
)

type MessagesService struct {
	repo        ports.MessagesRepo
	chatsRepo   ports.ChatsRepo
	usersRepo   ports.UsersRepo
	connManager *ConnectionsManager
}

func NewMessagesService(repo ports.MessagesRepo, chatsRepo ports.ChatsRepo) *MessagesService {
	return &MessagesService{
		repo:      repo,
		chatsRepo: chatsRepo,
	}
}

func (s *MessagesService) CreateMessage(ctx context.Context, m *models.Message) (err *errors.Error) {
	userId := auth.ExtractUser(ctx)
	ok, err := s.chatsRepo.CheckUserInChat(ctx, userId, m.ChatId)
	if err != nil {
		return err.Trace()
	}
	if !ok {
		return errors.New(fmt.Sprintf("user (%d) tried to create a message in the chat (%d)", userId, m.ChatId),
			models.ErrPermissionDenied, http.StatusForbidden)
	}
	m.UserId = userId
	m.Time = time.Now().UTC()

	ctx, err = db.WithTx(ctx, s.repo)
	if err != nil {
		return err.Trace()
	}
	defer db.CommitOnDefer(ctx, &err)

	if err := s.repo.New(ctx, m); err != nil {
		return err.Trace()
	}
	if err := s.chatsRepo.UpdateTime(ctx, m.ChatId, m.Time); err != nil {
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

	ctx, err = db.WithTx(ctx, s.repo)
	if err != nil {
		return err.Trace()
	}
	defer db.CommitOnDefer(ctx, &err)

	if err := s.repo.Delete(ctx, id); err != nil {
		return err.Trace()
	}
	if s.connManager != nil {
		go s.connManager.onDeleteMessage(id, m.ChatId)
	}

	lastMessId, err := s.repo.GetMinMassageIdInChat(ctx, m.ChatId)
	if err != nil {
		return err.Trace()
	}
	lastMess, err := s.repo.GetById(ctx, lastMessId)
	if err != nil {
		return err.Trace()
	}
	if err := s.chatsRepo.UpdateTime(ctx, m.ChatId, lastMess.Time); err != nil {
		return err.Trace()
	}

	return nil
}

func (s *MessagesService) GetFromChat(ctx context.Context, dto *GetMessagesDTO) ([]MessagesResponseDTO, *errors.Error) {
	if dto.Count <= 0 {
		return nil, errors.New1Msg("field count is missing", http.StatusBadRequest)
	}
	userId := auth.ExtractUser(ctx)
	ok, err := s.chatsRepo.CheckUserInChat(ctx, userId, dto.ChatId)
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
