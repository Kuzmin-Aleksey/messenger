package mysql

import "database/sql"

type Messages struct {
	db *sql.DB
}

func NewMessages(db *sql.DB) *Messages {
	return &Messages{db}
}
