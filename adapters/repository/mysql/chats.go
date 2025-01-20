package mysql

import "database/sql"

type Chats struct {
	db *sql.DB
}

func NewChats(db *sql.DB) *Chats {
	return &Chats{db}
}
