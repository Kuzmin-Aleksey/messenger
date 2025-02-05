package models

import (
	"database/sql"
	errorsutils "errors"
	"github.com/nyaruka/phonenumbers"
	"time"
)

type User struct {
	Id         int       `json:"id"`
	Phone      string    `json:"phone,omitempty"`
	Password   string    `json:"password,omitempty"`
	Name       string    `json:"name"`
	RealName   string    `json:"real_name"`
	ShowPhone  bool      `json:"show_phone"`
	LastOnline time.Time `json:"last_online"`
	Confirmed  bool      `json:"confirming"`
}

func (u *User) ScanRow(row *sql.Row) error {
	return row.Scan(
		&u.Id,
		&u.Phone,
		&u.Password,
		&u.Name,
		&u.RealName,
		&u.ShowPhone,
		&u.LastOnline,
		&u.Confirmed,
	)
}

func ParsePhone(phone string) (string, error) {
	num, err := phonenumbers.Parse(phone, "RU")
	if err != nil {
		return "", err
	}
	if !phonenumbers.IsValidNumber(num) {
		return "", errorsutils.New("invalid phone number")
	}
	return phonenumbers.Format(num, phonenumbers.E164), nil
}
