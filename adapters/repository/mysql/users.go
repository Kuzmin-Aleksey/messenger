package mysql

import (
	"database/sql"
	"errors"
	"messanger/models"
	tr "messanger/pkg/error_trace"
)

type Users struct {
	db *sql.DB
}

func NewUsers(db *sql.DB) *Users {
	return &Users{db}
}

func (u *Users) New(user *models.User) error {
	if _, err := u.db.Exec("INSERT INTO users (name) VALUES (?)", user.Name); err != nil {
		return tr.Trace(err)
	}
	return nil
}

const getUserChatsQuery = `
SELECT
	chats.id,
	chats.name
FROM user_2_chat
inner join chats on chats.id = user_2_chat.chat_id
WHERE user_2_chat.user_id = ?
`

func (u *Users) GetById(id int) (*models.User, error) {
	var user models.User
	if err := u.db.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(&user.Id, &user.Name); err != nil {
		return nil, tr.Trace(err)
	}

	rows, err := u.db.Query(getUserChatsQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &user, tr.Trace(err)
		}
		return nil, tr.Trace(err)
	}

	for rows.Next() {
		var chat models.Chat
		if err := rows.Scan(&chat.Id, &chat.Name); err != nil {
			return nil, tr.Trace(err)
		}
		user.Chats = append(user.Chats, chat)
	}

	return &user, nil
}

func (u *Users) AddChat(userId int, chatId int) error {
	if _, err := u.db.Exec("INSERT INTO user_2_chat (user_id, chat_id) VALUES (?, ?)", userId, chatId); err != nil {
		return tr.Trace(err)
	}
	return nil
}

func (u *Users) Delete(id int) error {
	if _, err := u.db.Exec("DELETE FROM users WHERE id = ?", id); err != nil {
		return tr.Trace(err)
	}

	if _, err := u.db.Exec("DELETE FROM user_2_chat WHERE user_id = ?", id); err != nil {
		return tr.Trace(err)
	}
	return nil

}
