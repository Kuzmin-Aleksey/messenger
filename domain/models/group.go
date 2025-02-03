package models

type Group struct {
	Id     int    `json:"id"`
	ChatId int    `json:"chat_id"`
	Name   string `json:"name"`
}
