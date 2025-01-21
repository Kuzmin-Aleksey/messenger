package mysql

import (
	"database/sql"
	"errors"
	"messanger/models"
	tr "messanger/pkg/error_trace"
)

type Chats struct {
	db *sql.DB
}

func NewChats(db *sql.DB) *Chats {
	return &Chats{db}
}

func (c *Chats) New(chat *models.Chat) error {
	res, err := c.db.Exec("INSERT INTO chats (name) VALUES (?)", chat.Name)
	if err != nil {
		return tr.Trace(err)
	}
	chatId, err := res.LastInsertId()
	if err != nil {
		return tr.Trace(err)
	}
	chat.Id = int(chatId)

	for _, user := range chat.Users {
		if _, err := c.db.Exec("INSERT INTO user_2_chat (user_id, chat_id) VALUES (?, ?)", user.Id, chat.Id); err != nil {
			return tr.Trace(err)
		}
	}

	return nil
}

func (c *Chats) Update(chat *models.Chat) error {
	if _, err := c.db.Exec("UPDATE chats SET name = ? WHERE id = ?", chat.Name, chat.Id); err != nil {
		return tr.Trace(err)
	}
	return nil
}

const getUsersByChatQuery = `
SELECT 
	users.id,
	users.name
FROM user_2_chat
inner join users ON users.id = user_2_chat.user_id
WHERE user_2_chat.chat_id = ?`

func (c *Chats) GetById(id int) (*models.Chat, error) {
	var chat models.Chat

	if err := c.db.QueryRow("SELECT * FROM chats WHERE id = ?", id).Scan(&chat.Id, &chat.Name); err != nil {
		return nil, tr.Trace(err)
	}

	rows, err := c.db.Query(getUsersByChatQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &chat, nil
		}
		return nil, tr.Trace(err)
	}

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.Id, &user.Name); err != nil {
			return nil, tr.Trace(err)
		}
		chat.Users = append(chat.Users, user)
	}

	return &chat, nil
}

func (c *Chats) AddUser(chatId int, userId int) error {
	if _, err := c.db.Exec("INSERT INTO user_2_chat (user_id, chat_id) VALUES (?, ?)", userId, chatId); err != nil {
		return tr.Trace(err)
	}
	return nil
}

func (c *Chats) CheckUserInChat(chatId int, userId int) (bool, error) {
	var exist int
	if err := c.db.QueryRow("SELECT COUNT(id) FROM user_2_chat WHERE user_id=? AND chat_id=?", userId, chatId).Scan(&exist); err != nil {
		return false, tr.Trace(err)
	}
	return exist != 0, nil
}

func (c *Chats) DeleteUser(chatId int, userId int) error {
	if _, err := c.db.Exec("DELETE FROM user_2_chat WHERE  user_id = ? AND chat_id = ?", userId, chatId); err != nil {
		return tr.Trace(err)
	}
	return nil
}

func (c *Chats) Delete(id int) error {
	if _, err := c.db.Exec("DELETE FROM chats WHERE id = ?", id); err != nil {
		return tr.Trace(err)
	}
	if _, err := c.db.Exec("DELETE FROM user_2_chat WHERE chat_id = ?", id); err != nil {
		return tr.Trace(err)
	}
	return nil
}
