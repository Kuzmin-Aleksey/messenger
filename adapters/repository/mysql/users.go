package mysql

import (
	"crypto/sha256"
	"database/sql"
	errorsutils "errors"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
)

type Users struct {
	db *sql.DB
}

func NewUsers(db *sql.DB) *Users {
	return &Users{db}
}

func (u *Users) New(user *domain.UserInfo) *errors.Error {
	if _, err := u.db.Exec("INSERT INTO users (id, name, email, password) VALUES (?, ?, ?, ?)",
		user.Id, user.Name, user.Email, hashPassword(user.Password)); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (u *Users) Update(user *domain.UserInfo) *errors.Error {
	if _, err := u.db.Exec("UPDATE users SET name = ? WHERE id = ?", user.Name, user.Id); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
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

func (u *Users) GetById(id int) (*domain.User, *errors.Error) {
	var user domain.User
	if err := u.db.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(&user.Id, &user.Name); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "user not found", http.StatusNotFound)
		}
		return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	rows, err := u.db.Query(getUserChatsQuery, id)
	if err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return &user, nil
		}
		return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	for rows.Next() {
		var chat domain.Chat
		if err := rows.Scan(&chat.Id, &chat.Name); err != nil {
			return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
		}
		user.Chats = append(user.Chats, chat)
	}

	return &user, nil
}

func (u *Users) GetInfoByEmail(email string) (*domain.UserInfo, *errors.Error) {
	var user domain.UserInfo

	if err := u.db.QueryRow("SELECT * FROM auth_users WHERE email = ?", email).Scan(
		&user.Id,
		&user.Email,
		&user.Password,
	); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "user not found", http.StatusNotFound)
		}
		return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	return &user, nil
}

func (u *Users) GetWithPass(email, password string) (*domain.UserInfo, *errors.Error) {
	user := new(domain.UserInfo)
	if err := u.db.QueryRow("SELECT * FROM users WHERE email = ? AND password = ?", email, hashPassword(password)).Scan(&user.Id, &user.Name, &user.Email, &user.Password); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "invalid email or password", http.StatusUnauthorized)
		}
		return nil, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	return user, nil
}

func (u *Users) Delete(id int) *errors.Error {
	tx, err := u.db.Begin()
	if err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	if _, err := tx.Exec("DELETE FROM users WHERE id = ?", id); err != nil {
		tx.Rollback()
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	if _, err := tx.Exec("DELETE FROM user_2_chat WHERE user_id = ?", id); err != nil {
		tx.Rollback()
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	if err := tx.Commit(); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

var passwordSalt = []byte("nwy9837ctp5")

func hashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return string(h.Sum(passwordSalt))
}
