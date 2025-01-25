package mysql

import (
	"crypto/sha256"
	"database/sql"
	errorsutils "errors"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
	"strings"
)

type Users struct {
	db *sql.DB
}

func NewUsers(db *sql.DB) *Users {
	return &Users{db}
}

func (u *Users) New(user *domain.User) *errors.Error {
	res, err := u.db.Exec("INSERT INTO users (name, email, password) VALUES (?, ?, ?)",
		user.Name, user.Email, hashPassword(user.Password))
	if err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	userId, _ := res.LastInsertId()
	user.Id = int(userId)
	return nil
}

func (u *Users) SetConfirm(userId int, v bool) *errors.Error {
	if _, err := u.db.Exec("UPDATE users SET confirmed = ? WHERE id = ?", v, userId); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (u *Users) AddUserToChat(userId int, chatId int, role string) *errors.Error {
	var roleId int
	if err := u.db.QueryRow("SELECT id FROM roles WHERE role=?", role).Scan(&roleId); err != nil {
		return errors.New(err, "unknown role", http.StatusConflict)
	}
	if _, err := u.db.Exec("INSERT INTO user_2_chat (user_id, chat_id, role_id) VALUES (?, ?, ?)",
		userId, chatId, roleId); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (u *Users) CheckUserInChat(userId int, chatId int) (bool, *errors.Error) {
	var exist bool
	if err := u.db.QueryRow("SELECT COUNT(id) > 0 FROM user_2_chat WHERE user_id=? AND chat_id=?", userId, chatId).Scan(&exist); err != nil {
		return false, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return exist, nil
}

func (u *Users) DeleteUserFromChat(chatId int, userId int) *errors.Error {
	if _, err := u.db.Exec("DELETE FROM user_2_chat WHERE user_id = ? AND chat_id = ?", userId, chatId); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

const getRoleQuery = `
SELECT roles.role
FROM user_2_chat
INNER JOIN roles ON user_2_chat.role_id = roles.id
WHERE user_2_chat.user_id=? AND user_2_chat.chat_id=?`

func (u *Users) GetRole(userId int, chatId int) (string, *errors.Error) {
	var role string
	if err := u.db.QueryRow(getRoleQuery, userId, chatId).Scan(&role); err != nil {
		return "", errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return role, nil
}

const getUsersByChatQuery = `
SELECT 
	users.id,
	users.name,
	IFNULL(roles.role, 
	(SELECT r.role FROM roles r WHERE r.id=1)) AS role
FROM user_2_chat
INNER JOIN users ON users.id = user_2_chat.user_id
LEFT JOIN roles ON roles.id = user_2_chat.role_id
WHERE user_2_chat.chat_id = ?`

func (u *Users) GetUsersByChat(chatId int) ([]domain.User, *errors.Error) {
	rows, err := u.db.Query(getUsersByChatQuery, chatId)
	if err != nil {
		return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	var users []domain.User

	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.Id, &user.Name, &user.Role); err != nil {
			return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
		}
		users = append(users, user)
	}
	return users, nil
}

func (u *Users) CountOfUsersInChat(chatId int) (int, *errors.Error) {
	var count int
	if err := u.db.QueryRow("SELECT COUNT(id) FROM user_2_chat WHERE chat_id=?", chatId).Scan(&count); err != nil {
		return 0, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return count, nil
}

func (u *Users) Update(user *domain.User) *errors.Error {
	query := "UPDATE users SET"
	var values []any
	if user.Email != "" {
		query += " email = ?,"
		values = append(values, user.Email)
	}
	if user.Name != "" {
		query += " name = ?,"
		values = append(values, user.Name)
	}
	if user.Password != "" {
		query += " password = ?,"
		values = append(values, hashPassword(user.Password))
	}
	if len(values) == 0 {
		return nil
	}
	query = strings.TrimSuffix(query, ",")
	query += " WHERE id = ?"
	values = append(values, user.Id)

	if _, err := u.db.Exec(query, values...); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (u *Users) GetById(id int) (*domain.User, *errors.Error) {
	var user domain.User
	if err := u.db.QueryRow("SELECT id, name, email FROM users WHERE id = ?", id).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
	); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "user not found", http.StatusNotFound)
		}
		return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return &user, nil
}

func (u *Users) GetByEmail(email string) (*domain.User, *errors.Error) {
	var user domain.User
	if err := u.db.QueryRow("SELECT id, name, email FROM users WHERE email = ?", email).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
	); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "user not found", http.StatusNotFound)
		}
		return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return &user, nil
}

func (u *Users) GetByEmailWithPass(email, password string) (*domain.User, *errors.Error) {
	user := new(domain.User)
	if err := u.db.QueryRow("SELECT id, name, email  FROM users WHERE email = ? AND password = ? AND confirmed", email, hashPassword(password)).Scan(
		&user.Id,
		&user.Name,
		&user.Email); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "invalid email or password", http.StatusUnauthorized)
		}
		return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return user, nil
}

func (u *Users) CheckPass(id int, password string) (bool, *errors.Error) {
	var ok bool
	if err := u.db.QueryRow("SELECT COUNT(id) > 0  FROM users WHERE id = ? AND password = ?", id, hashPassword(password)).Scan(&ok); err != nil {
		return false, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return ok, nil
}

func (u *Users) Delete(id int) (e *errors.Error) {
	tx, err := u.db.Begin()
	if err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	defer Commit(tx, &e)

	if _, err := tx.Exec("DELETE FROM users WHERE id = ?", id); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	if _, err := tx.Exec("DELETE FROM user_2_chat WHERE user_id = ?", id); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

var passwordSalt = []byte("nwy9837ctp5")

func hashPassword(password string) []byte {
	h := sha256.New()
	h.Write([]byte(password))
	return h.Sum(passwordSalt)
}
