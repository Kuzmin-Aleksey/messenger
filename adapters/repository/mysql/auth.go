package mysql

import (
	"database/sql"
	"messanger/models"
	tr "messanger/pkg/error_trace"
)

type AuthUsers struct {
	db *sql.DB
}

func NewAuthUsers(db *sql.DB) *AuthUsers {
	return &AuthUsers{db}
}

func (u *AuthUsers) New(user *models.AuthUser) error {
	res, err := u.db.Exec("INSERT INTO auth_users (email, password) VALUE (?, ?)")
	if err != nil {
		return tr.Trace(err)
	}

	userId, err := res.LastInsertId()
	if err != nil {
		return tr.Trace(err)
	}
	user.Id = int(userId)
	return nil
}

func (u *AuthUsers) GetUserByEmail(email string) (*models.AuthUser, error) {
	var user models.AuthUser

	if err := u.db.QueryRow("SELECT * FROM auth_users WHERE email = ?", email).Scan(
		&user.Id,
		&user.Email,
		&user.Password,
	); err != nil {
		return nil, tr.Trace(err)
	}

	return &user, nil
}

func (u *AuthUsers) CheckPassword(password string, email string) (bool, error) {
	var exist int

	if err := u.db.QueryRow("SELECT COUNT(id) FROM auth_users WHERE email = ? AND password = ?", email, password).Scan(&exist); err != nil {
		return false, tr.Trace(err)
	}

	return exist != 0, nil
}

func (u *AuthUsers) IsExist(email string) (bool, error) {
	var exist int

	if err := u.db.QueryRow("SELECT COUNT(id) FROM auth_users WHERE email = ?", email).Scan(&exist); err != nil {
		return false, tr.Trace(err)
	}

	return exist != 0, nil
}

func (u *AuthUsers) Delete(is int) error {
	if _, err := u.db.Exec("DELETE FROM auth_users WHERE id = ?", is); err != nil {
		return tr.Trace(err)
	}
	return nil
}
