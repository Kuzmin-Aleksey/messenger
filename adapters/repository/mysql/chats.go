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

func (c *Chats) New(chat *domain.Chat, creator int) (e *errors.Error) {
	tx, err := c.db.Begin()
	if err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	defer Commit(tx, &e)

	res, err := tx.Exec("INSERT INTO chats (name) VALUES (?)", chat.Name)
	if err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	chatId, err := res.LastInsertId()
	if err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	chat.Id = int(chatId)

	if _, err := tx.Exec("INSERT INTO user_2_chat (user_id, chat_id, role_id) VALUES (?, ?, 2)", creator, chatId); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	return nil
}

func (c *Chats) Update(chat *domain.Chat) *errors.Error {
	if len(chat.Name) == 0 {
		return nil
	}
	if _, err := c.db.Exec("UPDATE chats SET name = ? WHERE id = ?", chat.Name, chat.Id); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

const getChatsByUserQuery = `
SELECT
	chats.id,
	chats.name,
	IFNULL(roles.role, 
	(SELECT r.role FROM roles r WHERE r.id=1)) AS role
FROM user_2_chat
INNER JOIN chats ON chats.id = user_2_chat.chat_id
LEFT JOIN roles ON roles.id = user_2_chat.role_id
WHERE user_2_chat.user_id = ?
`

func (c *Chats) GetChatsByUser(userId int) ([]domain.Chat, *errors.Error) {
	rows, err := c.db.Query(getChatsByUserQuery, userId)
	if err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return []domain.Chat{}, nil
		}
		return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	var chats []domain.Chat
	for rows.Next() {
		var chat domain.Chat
		if err := rows.Scan(&chat.Id, &chat.Name, new(string)); err != nil {
			return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
		}
		chats = append(chats, chat)
	}
	return chats, nil
}

func (c *Chats) GetById(id int) (*domain.Chat, *errors.Error) {
	var chat domain.Chat
	if err := c.db.QueryRow("SELECT * FROM chats WHERE id = ?", id).Scan(&chat.Id, &chat.Name); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "chat not found", http.StatusNotFound)
		}
		return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return &chat, nil
}

const getUserRoleQuery = `
SELECT roles.role
FROM user_2_chat
INNER JOIN roles ON user_2_chat.role_id = roles.id
WHERE user_2_chat.user_id=? AND user_2_chat.chat_id=?`

func (c *Chats) GetUserRole(userId int, chatId int) (string, *errors.Error) {
	var role string
	if err := c.db.QueryRow(getUserRoleQuery, userId, chatId).Scan(&role); err != nil {
		return "", errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return role, nil
}

func (c *Chats) Delete(id int) (e *errors.Error) {
	tx, err := c.db.Begin()
	if err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	defer Commit(tx, &e)

	if _, err := c.db.Exec("DELETE FROM chats WHERE id = ?", id); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	if _, err := c.db.Exec("DELETE FROM user_2_chat WHERE chat_id = ?", id); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}
