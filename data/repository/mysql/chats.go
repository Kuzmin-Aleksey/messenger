package mysql

import (
	"context"
	"database/sql"
	errorsutils "errors"
	"messanger/domain/models"
	"messanger/pkg/errors"
	"net/http"
)

type Chats struct {
	db *sql.DB
}

func NewChats(db *sql.DB) *Chats {
	return &Chats{db: db}
}

func (c *Chats) New(ctx context.Context, chat *models.Chat, users []int) (e *errors.Error) {
	res, err := c.db.ExecContext(ctx, "INSERT INTO chats (type) VALUE (?)", chat.Type)
	if err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	chatId, err := res.LastInsertId()
	if err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	chat.Id = int(chatId)

	for _, userId := range users {
		if err := c.AddUserToChat(ctx, chat.Id, userId); err != nil {
			return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
		}
	}

	return nil
}

func (c *Chats) AddUserToChat(ctx context.Context, id int, userId int) *errors.Error {
	if _, err := c.db.ExecContext(ctx, "INSERT INTO user_2_chat (user_id, chat_id) VALUES (?, ?)", userId, id); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (c *Chats) RemoveUserFromChat(ctx context.Context, id int, userId int) *errors.Error {
	if _, err := c.db.ExecContext(ctx, "DELETE FROM user_2_chat WHERE user_id=? AND chat_id=?", userId, id); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (c *Chats) CheckUserInChat(ctx context.Context, userId int, chatId int) (bool, *errors.Error) {
	var exist bool
	if err := c.db.QueryRowContext(ctx, "SELECT COUNT(id) != 0 as exist FROM user_2_chat WHERE user_id=? AND chat_id=?", userId, chatId).Scan(&exist); err != nil {
		return exist, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return exist, nil
}

func (c *Chats) CountUsersInChat(ctx context.Context, id int) (int, *errors.Error) {
	var count int
	if err := c.db.QueryRowContext(ctx, "SELECT COUNT(id) as count FROM user_2_chat WHERE chat_id=?", id).Scan(&count); err != nil {
		return 0, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return count, nil
}

const getChatsByUserQuery = `
SELECT 
chats.id,
chats.type
FROM user_2_chat
INNER JOIN chats ON user_2_chat.chat_id = chats.id
WHERE user_2_chat.user_id = ?
`

func (c *Chats) GetByUserId(ctx context.Context, userId int) ([]models.Chat, *errors.Error) {
	chats := make([]models.Chat, 0)
	rows, err := c.db.QueryContext(ctx, getChatsByUserQuery, userId)
	if err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return chats, nil
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	defer rows.Close()

	for rows.Next() {
		var chat models.Chat
		if err := rows.Scan(&chat.Id, &chat.Type); err != nil {
			return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
		}
		chats = append(chats, chat)
	}
	return chats, nil
}

func (c *Chats) GetChatListByUser(ctx context.Context, userId int) ([]int, *errors.Error) {
	rows, err := c.db.QueryContext(ctx, "SELECT chat_id FROM user_2_chat WHERE user_id = ?", userId)
	if err != nil {
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	defer rows.Close()

	var chatIds []int
	for rows.Next() {
		var chatId int
		if err := rows.Scan(&chatId); err != nil {
			return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
		}
		chatIds = append(chatIds, chatId)
	}
	return chatIds, nil
}

const getUserCompanionByChatIdQuery = `
SELECT
uc.user_id
FROM user_2_chat as uc
INNER JOIN chats ON chats.id = uc.chat_id
WHERE 
	uc.chat_id = ? AND
    uc.user_id != ?;
`

func (c *Chats) GetUserCompanionByChatId(ctx context.Context, userId int, chatId int) (int, *errors.Error) {
	var respUserId int
	if err := c.db.QueryRowContext(ctx, getUserCompanionByChatIdQuery, chatId, userId).Scan(&respUserId); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return 0, errors.New(err, "user not found", http.StatusNotFound)
		}
		return 0, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return respUserId, nil
}

func (c *Chats) GetById(ctx context.Context, id int) (*models.Chat, *errors.Error) {
	chat := new(models.Chat)
	if err := c.db.QueryRowContext(ctx, "SELECT * FROM chats WHERE id=?", id).Scan(&chat.Id, &chat.Type, &chat.CreateTime); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return chat, errors.New(err, "chat not found", http.StatusNotFound)
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return chat, nil
}

func (c *Chats) Delete(ctx context.Context, id int) (e *errors.Error) {
	if _, err := c.db.ExecContext(ctx, "DELETE FROM chats WHERE id = ?", id); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}
