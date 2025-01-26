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

func (m *Messages) GetByChat(chatId int, lastId int, count int) ([]domain.Message, *errors.Error) {
	messages := make([]domain.Message, 0)

	var query string
	var args []any
	if lastId > 0 {
		query = "SELECT * FROM messages WHERE chat_id = ? AND id < ? ORDER BY time DESC LIMIT ?"
		args = []any{chatId, lastId, count}
	} else {
		query = `SELECT * FROM messages WHERE chat_id = ? ORDER BY time DESC LIMIT ?`
		args = []any{chatId, count}
	}

	rows, err := m.db.Query(query, args...)
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
		messages = append(messages, message)
	}
	return messages, nil
}

func (m *Messages) GetMinMassageIdInChat(chatId int) (int, *errors.Error) {
	var id int
	if err := m.db.QueryRow("SELECT IFNULL(MIN(id), -1) FROM messages WHERE chat_id = ?", chatId).Scan(&id); err != nil {
		return id, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return id, nil
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
