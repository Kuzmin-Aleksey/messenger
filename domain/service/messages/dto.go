package messages

import "time"

type GetMessagesDTO struct {
	ChatId        int `json:"chat_id"`
	LastMessageId int `json:"last_message_id"`
	Count         int `json:"count"`
}

type MessagesResponseDTO struct {
	Id     int       `json:"id"`
	UserId int       `json:"user_id"`
	Text   string    `json:"text"`
	Time   time.Time `json:"time"`
}

type UpdateMessageDTO struct {
	Id   int    `json:"id"`
	Text string `json:"text"`
}
