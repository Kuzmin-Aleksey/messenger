package domain

import "time"

type Message struct {
	Id     int       `json:"id"`
	ChatId int       `json:"chat_id"`
	UserId int       `json:"user_id"`
	Text   string    `json:"text"`
	Time   time.Time `json:"time"`
}
