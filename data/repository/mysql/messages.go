package mysql

import (
	"context"
	"database/sql"
	errorsutils "errors"
	"messanger/domain/models"
	"messanger/pkg/errors"
	"net/http"
)

type Messages struct {
	db *sql.DB
}

func NewMessages(db *sql.DB) *Messages {
	return &Messages{db}
}

func (m *Messages) New(ctx context.Context, message *models.Message) *errors.Error {
	res, err := m.db.Exec("INSERT INTO messages (chat_id, user_id, value, time) VALUE (?, ?, ?, ?)",
		message.ChatId, message.UserId, message.Text, message.Time)
	if err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	message.Id = int(id)
	return nil
}

func (m *Messages) GetByChat(ctx context.Context, chatId int, lastId int, count int) ([]models.Message, *errors.Error) {
	messages := make([]models.Message, 0)

	var query string
	var args []any
	if lastId > 0 {
		query = "SELECT * FROM messages WHERE chat_id = ? AND id < ? ORDER BY time DESC LIMIT ?"
		args = []any{chatId, lastId, count}
	} else {
		query = `SELECT * FROM messages WHERE chat_id = ? ORDER BY time DESC LIMIT ?`
		args = []any{chatId, count}
	}

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return messages, nil
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}

	for rows.Next() {
		var message models.Message
		if err := rows.Scan(&message.Id, &message.ChatId, &message.UserId, &message.Text, &message.Time); err != nil {
			return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func (m *Messages) GetMinMassageIdInChat(ctx context.Context, chatId int) (int, *errors.Error) {
	var id int
	if err := m.db.QueryRowContext(ctx, "SELECT IFNULL(MIN(id), -1) FROM messages WHERE chat_id = ?", chatId).Scan(&id); err != nil {
		return id, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return id, nil
}

func (m *Messages) GetLastMessage(ctx context.Context, chatId int) (*models.Message, *errors.Error) {
	var message models.Message
	if err := m.db.QueryRowContext(ctx, "SELECT * FROM messages WHERE chat_id = ? ORDER BY time DESC", chatId).Scan(
		&message.Id, &message.ChatId, &message.UserId, &message.Text, &message.Time); err != nil {
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return &message, nil
}

func (m *Messages) IsUserMessage(ctx context.Context, id int, userId int) (bool, *errors.Error) {
	var ok bool
	if err := m.db.QueryRowContext(ctx, "SELECT user_id=? FROM messages WHERE id=?", userId, id).Scan(&ok); err != nil {
		return false, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return ok, nil
}

func (m *Messages) GetById(ctx context.Context, id int) (*models.Message, *errors.Error) {
	var message models.Message
	if err := m.db.QueryRowContext(ctx, "SELECT * FROM messages WHERE id = ?", id).Scan(
		&message.Id, &message.ChatId, &message.UserId, &message.Text, &message.Time); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "message not found", http.StatusNotFound)
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return &message, nil
}

func (m *Messages) Update(ctx context.Context, id int, text string) *errors.Error {
	if _, err := m.db.ExecContext(ctx, "UPDATE messages SET value=? WHERE id = ?", text, id); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (m *Messages) Delete(ctx context.Context, id int) *errors.Error {
	if _, err := m.db.ExecContext(ctx, "DELETE FROM messages WHERE id = ?", id); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}
