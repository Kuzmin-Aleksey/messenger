package models

import (
	"database/sql"
	"time"
)

type User struct {
	Id             int       `json:"id"`
	Phone          string    `json:"phone,omitempty"`
	Password       string    `json:"password,omitempty"`
	Name           string    `json:"name"`
	RealName       string    `json:"real_name"`
	CanFindByPhone bool      `json:"can_find_by_phone"`
	LastOnline     time.Time `json:"last_online"`
	Confirmed      bool      `json:"confirming"`
}

func (u *User) ScanRow(row *sql.Row) error {
	return row.Scan(
		&u.Id,
		&u.Phone,
		&u.Password,
		&u.Name,
		&u.RealName,
		&u.CanFindByPhone,
		&u.LastOnline,
		&u.Confirmed,
	)
}
