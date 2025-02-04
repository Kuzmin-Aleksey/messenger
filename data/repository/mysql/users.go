package mysql

import (
	"context"
	"crypto/sha256"
	"database/sql"
	errorsutils "errors"
	"messanger/domain/models"
	"messanger/pkg/errors"
	"net/http"
	"time"
)

type Users struct {
	DB
}

func NewUsers(db DB) *Users {
	return &Users{db}
}

func (u *Users) New(ctx context.Context, user *models.User) *errors.Error {
	res, err := u.DB.ExecContext(ctx, "INSERT INTO users (phone, password, name, real_namel) VALUES (?, ?, ?, ?)",
		user.Phone, hashPassword(user.Password), user.Name, user.RealName)

	if err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	userId, _ := res.LastInsertId()
	user.Id = int(userId)
	return nil
}

func (u *Users) SetConfirm(ctx context.Context, userId int, v bool) *errors.Error {
	if _, err := u.DB.ExecContext(ctx, "UPDATE users SET confirmed = ? WHERE id = ?", v, userId); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (u *Users) UpdateUsername(ctx context.Context, userId int, name string) *errors.Error {
	if _, err := u.DB.ExecContext(ctx, "UPDATE users SET name = ? WHERE id = ?", name, userId); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (u *Users) UpdateRealName(ctx context.Context, userId int, realName string) *errors.Error {
	if _, err := u.DB.ExecContext(ctx, "UPDATE users SET real_namel = ? WHERE id = ?", realName, userId); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (u *Users) UpdatePassword(ctx context.Context, userId int, password string) *errors.Error {
	if _, err := u.DB.ExecContext(ctx, "UPDATE users SET password = ? WHERE id = ?", hashPassword(password), userId); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (u *Users) UpdatePhone(ctx context.Context, userId int, phone string) *errors.Error {
	if _, err := u.DB.ExecContext(ctx, "UPDATE users SET phone = ? WHERE id = ?", phone, userId); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (u *Users) UpdateLastOnlineTime(ctx context.Context, userId int, time time.Time) *errors.Error {
	if _, err := u.DB.ExecContext(ctx, "UPDATE users SET last_online = ? WHERE id = ?", time, userId); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (u *Users) GetById(ctx context.Context, id int) (*models.User, *errors.Error) {
	var user models.User
	if err := user.ScanRow(u.DB.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?", id)); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "user not found", http.StatusNotFound)
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return &user, nil
}

func (u *Users) FindByPhone(ctx context.Context, phone string) (*models.User, *errors.Error) {
	var user models.User
	if err := user.ScanRow(u.DB.QueryRowContext(ctx, "SELECT * FROM users WHERE phone = ?", phone)); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "user not found", http.StatusNotFound)
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return &user, nil
}

func (u *Users) FindByName(ctx context.Context, name string) (*models.User, *errors.Error) {
	var user models.User
	if err := user.ScanRow(u.DB.QueryRowContext(ctx, "SELECT * FROM users WHERE name = ?", name)); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "user not found", http.StatusNotFound)
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return &user, nil
}

func (u *Users) GetByPhoneWithPass(ctx context.Context, phone, password string) (*models.User, *errors.Error) {
	user := new(models.User)
	if err := user.ScanRow(u.DB.QueryRowContext(ctx, "SELECT * FROM users WHERE phone = ? AND password = ? AND confirmed",
		phone, hashPassword(password))); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "invalid phone or password", http.StatusUnauthorized)
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return user, nil
}

func (u *Users) GetByIdWithPass(ctx context.Context, id int, password string) (*models.User, *errors.Error) {
	user := new(models.User)
	if err := user.ScanRow(u.DB.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ? AND password = ? AND confirmed",
		id, hashPassword(password))); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return nil, errors.New(err, "invalid password", http.StatusUnauthorized)
		}
		return nil, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return user, nil
}

func (u *Users) Delete(ctx context.Context, id int) (e *errors.Error) {
	if _, err := u.DB.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	if _, err := u.DB.ExecContext(ctx, "DELETE FROM user_2_chat WHERE user_id = ?", id); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (u *Users) GetLastOnline(ctx context.Context, userId int) (time.Time, *errors.Error) {
	var t time.Time
	if err := u.DB.QueryRowContext(ctx, "SELECT last_online FROM users WHERE id = ?", userId).Scan(&t); err != nil {
		if errorsutils.Is(err, sql.ErrNoRows) {
			return t, errors.New(err, "user not found", http.StatusNotFound)
		}
		return t, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return t, nil
}

var passwordSalt = []byte("nwy9837ctp5")

func hashPassword(password string) []byte {
	h := sha256.New()
	h.Write([]byte(password))
	return h.Sum(passwordSalt)
}
