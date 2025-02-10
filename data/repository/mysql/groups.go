package mysql

import (
	"context"
	"database/sql"
	errorsutils "errors"
	"fmt"
	"messanger/domain/models"
	"messanger/pkg/errors"
	"net/http"
	"time"
)

type Groups struct {
	DB
}

func NewGroups(db DB) (*Groups, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := openAndExec(ctx, db, "data/repository/mysql/scripts/create_groups.sql"); err != nil {
		return nil, errorsutils.New("create table groups error: " + err.Error())
	}
	if err := openAndExec(ctx, db, "data/repository/mysql/scripts/create_roles.sql"); err != nil {
		return nil, errorsutils.New("create table roles error: " + err.Error())
	}

	return &Groups{DB: db}, nil
}

func (g *Groups) New(ctx context.Context, group *models.Group) *errors.Error {
	res, err := g.DB.ExecContext(ctx, "INSERT INTO `groups` (chat_id, name) VALUES (?, ?)", group.ChatId, group.Name)
	if err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	groupId, err := res.LastInsertId()
	if err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	group.Id = int(groupId)
	return nil
}

func (g *Groups) Update(ctx context.Context, group *models.Group) *errors.Error {
	if len(group.Name) == 0 {
		return nil
	}
	if _, err := g.DB.ExecContext(ctx, "UPDATE `groups` SET name = ? WHERE id = ?", group.Name, group.Id); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

const getGroupsByUserQuery = "SELECT g.id, g.chat_id, g.name FROM user_2_chat INNER JOIN `groups` g ON g.chat_id = user_2_chat.chat_id WHERE user_2_chat.user_id = ?"

func (g *Groups) GetGroupsByUser(ctx context.Context, userId int) ([]models.Group, *errors.Error) {
	rows, err := g.DB.QueryContext(ctx, getGroupsByUserQuery, userId)
	if err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return []models.Group{}, nil
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.Id, &group.ChatId, &group.Name); err != nil {
			return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (g *Groups) GetGroupByChatId(ctx context.Context, chatId int) (*models.Group, *errors.Error) {
	var group models.Group
	if err := g.DB.QueryRowContext(ctx, "SELECT * FROM `groups` WHERE chat_id = ?", chatId).Scan(&group.Id, &group.ChatId, &group.Name); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, fmt.Sprintf("group not found by chat id = %d", chatId), http.StatusNotFound)
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return &group, nil
}

const getUsersByGroupQuery = `
SELECT 
 user_2_chat.user_id
FROM user_2_chat
INNER JOIN ` + "`groups`" + ` g ON g.chat_id = user_2_chat.chat_id
WHERE g.id = ?
`

func (g *Groups) GetUsersByGroup(ctx context.Context, id int) ([]int, *errors.Error) {
	var users []int
	rows, err := g.DB.QueryContext(ctx, getUsersByGroupQuery, id)
	if err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return users, nil
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
		}
		users = append(users, id)
	}
	return users, nil
}

const checkUserInGroupQuery = `
SELECT
    COUNT(*) != 0 AS exist
FROM user_2_chat
INNER JOIN ` + "`groups`" + ` g on user_2_chat.chat_id = g.chat_id
WHERE 
    user_2_chat.user_id = ? AND
    g.id = ?
`

func (g *Groups) CheckUserInGroup(ctx context.Context, userId int, groupId int) (bool, *errors.Error) {
	var exist bool
	if err := g.DB.QueryRowContext(ctx, checkUserInGroupQuery, userId, groupId).Scan(&exist); err != nil {
		return false, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return exist, nil
}

func (g *Groups) GetById(ctx context.Context, id int) (*models.Group, *errors.Error) {
	var group models.Group
	if err := g.DB.QueryRowContext(ctx, "SELECT * FROM `groups` WHERE id = ?", id).Scan(&group.Id, &group.ChatId, &group.Name); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "group not found", http.StatusNotFound)
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return &group, nil
}

func (g *Groups) SetRole(ctx context.Context, userId int, groupId int, role string) *errors.Error {
	var existRole string
	if err := g.DB.QueryRowContext(ctx, "SELECT role FROM roles WHERE user_id=? AND group_id=?", userId, groupId).Scan(&existRole); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			if _, err := g.DB.ExecContext(ctx, "INSERT INTO roles (user_id, group_id, role) VALUES (?, ?, ?)", userId, groupId, role); err != nil {
				return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
			}
			return nil
		}
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	if existRole != role {
		if _, err := g.DB.ExecContext(ctx, "UPDATE roles SET role=? WHERE user_id=? AND group_id=?", role, userId, groupId); err != nil {
			return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
		}
	}
	return nil
}

func (g *Groups) GetRole(ctx context.Context, userId int, groupId int) (string, *errors.Error) {
	var role string
	if err := g.DB.QueryRowContext(ctx, "SELECT role FROM roles WHERE user_id=? AND group_id=?", userId, groupId).Scan(&role); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return "", errors.New(err, "role not found", http.StatusNotFound)
		}
		return "", errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return role, nil
}

func (g *Groups) Delete(ctx context.Context, id int) (e *errors.Error) {
	if _, err := g.DB.ExecContext(ctx, "DELETE FROM `groups` WHERE id = ?", id); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}
