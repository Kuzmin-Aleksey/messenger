package models

import "time"

type Chat struct {
	Id              int       `json:"id"`
	Type            string    `json:"type"`
	CreateTime      time.Time `json:"create_time"`
	LastMessageTime time.Time `json:"last_message_time"`
}

const (
	ChatTypeGroup = "group"
	ChatTypeUser  = "user"
)
