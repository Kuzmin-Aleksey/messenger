package mysql

import (
	"database/sql"
	errorsutils "errors"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
)

type Messages struct {
	db *sql.DB
}

func NewMessages(db *sql.DB) *Messages {
	return &Messages{db}
}

func (m *Messages) New(message *domain.Message) *errors.Error {
	res, err := m.db.Exec("INSERT INTO messages (chat_id, user_id, value, time) VALUE (?, ?, ?, ?)",
		message.ChatId, message.UserId, message.Text, message.Time)
	if err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	message.Id = int(id)
	return nil
}

const getMessagesByChatQuery = `
SELECT * FROM messages
WHERE chat_id = ? AND id < ?
ORDER BY time DESC
LIMIT ?;
`

func (m *Messages) GetByChat(chatId int, lastId int, count int) ([]domain.Message, *errors.Error) {
	var messages []domain.Message
	rows, err := m.db.Query(getMessagesByChatQuery, chatId, lastId, count)
	if err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return messages, nil
		}
		return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	for rows.Next() {
		var message domain.Message
		if err := rows.Scan(&message.Id, &message.ChatId, &message.UserId, &message.Text, &message.Time); err != nil {
			return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
		}
	}
	return messages, nil
}

func (m *Messages) GetById(id int) (*domain.Message, *errors.Error) {
	var message domain.Message
	if err := m.db.QueryRow("SELECT * FROM messages WHERE id = ?", id).Scan(
		&message.Id, &message.ChatId, &message.UserId, &message.Text, &message.Time); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "message not found", http.StatusNotFound)
		}
		return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return &message, nil
}

func (m *Messages) Update(id int, text string) *errors.Error {
	if _, err := m.db.Exec("UPDATE messages SET value=? WHERE id = ?", text, id); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (m *Messages) Delete(id int) *errors.Error {
	if _, err := m.db.Exec("DELETE FROM messages WHERE id = ?", id); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (m *Messages) DeleteByChat(chatId int) *errors.Error {
	if _, err := m.db.Exec("DELETE FROM messages WHERE chat_id = ?", chatId); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}
