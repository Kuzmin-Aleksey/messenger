package mysql

import (
	"database/sql"
	errorsutils "errors"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
)

type Chats struct {
	db *sql.DB
}

func NewChats(db *sql.DB) *Chats {
	return &Chats{db}
}

func (c *Chats) New(chat *domain.Chat) *errors.Error {
	res, err := c.db.Exec("INSERT INTO chats (name) VALUES (?)", chat.Name)
	if err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	chatId, err := res.LastInsertId()
	if err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	chat.Id = int(chatId)

	for _, user := range chat.Users {
		if _, err := c.db.Exec("INSERT INTO user_2_chat (user_id, chat_id) VALUES (?, ?)", user.Id, chat.Id); err != nil {
			return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
		}
	}

	return nil
}

func (c *Chats) Update(chat *domain.Chat) *errors.Error {
	if _, err := c.db.Exec("UPDATE chats SET name = ? WHERE id = ?", chat.Name, chat.Id); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
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

func (c *Chats) GetById(id int) (*domain.Chat, *errors.Error) {
	var chat domain.Chat

	if err := c.db.QueryRow("SELECT * FROM chats WHERE id = ?", id).Scan(&chat.Id, &chat.Name); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "chat not found", http.StatusNotFound)
		}
		return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	rows, err := c.db.Query(getUsersByChatQuery, id)
	if err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return &chat, nil
		}
		return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.Id, &user.Name); err != nil {
			return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
		}
		chat.Users = append(chat.Users, user)
	}

	return &chat, nil
}

func (c *Chats) AddUser(chatId int, userId int) *errors.Error {
	if _, err := c.db.Exec("INSERT INTO user_2_chat (user_id, chat_id) VALUES (?, ?)", userId, chatId); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (c *Chats) CheckUserInChat(chatId int, userId int) (bool, *errors.Error) {
	var exist int
	if err := c.db.QueryRow("SELECT COUNT(id) FROM user_2_chat WHERE user_id=? AND chat_id=?", userId, chatId).Scan(&exist); err != nil {
		return false, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return exist != 0, nil
}

func (c *Chats) DeleteUser(chatId int, userId int) *errors.Error {
	if _, err := c.db.Exec("DELETE FROM user_2_chat WHERE  user_id = ? AND chat_id = ?", userId, chatId); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (c *Chats) Delete(id int) *errors.Error {
	tx, err := c.db.Begin()
	if err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	if _, err := c.db.Exec("DELETE FROM chats WHERE id = ?", id); err != nil {
		tx.Rollback()
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	if _, err := c.db.Exec("DELETE FROM user_2_chat WHERE chat_id = ?", id); err != nil {
		tx.Rollback()
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	if err := tx.Commit(); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}
