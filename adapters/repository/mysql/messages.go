package mysql

import (
	"database/sql"
	"errors"
	"messanger/models"
	tr "messanger/pkg/error_trace"
)

type Messages struct {
	db *sql.DB
}

func NewMessages(db *sql.DB) *Messages {
	return &Messages{db}
}

func (m *Messages) New(message *models.Message) error {
	res, err := m.db.Exec("INSERT INTO messages (chat_id, user_id, value, time) VALUE (?, ?, ?, ?)",
		message.ChatId, message.UserId, message.Text, message.Time)
	if err != nil {
		return tr.Trace(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return tr.Trace(err)
	}
	message.Id = int(id)

	return nil
}

func (m *Messages) GetByChat(chatId int, offset int, count int) ([]models.Message, error) {
	var messages []models.Message
	rows, err := m.db.Query("SELECT * FROM messages WHERE chat_id = ? ORDER BY time LIMIT ?, ?", chatId, offset, count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return messages, nil
		}
	}

	for rows.Next() {
		var message models.Message
		if err := rows.Scan(&message.Id, &message.ChatId, &message.UserId, &message.Text, &message.Time); err != nil {
			return nil, tr.Trace(err)
		}
	}
	return messages, nil
}

func (m *Messages) Delete(id int) error {
	if _, err := m.db.Exec("DELETE FROM messages WHERE id = ?", id); err != nil {
		return tr.Trace(err)
	}
	return nil
}
