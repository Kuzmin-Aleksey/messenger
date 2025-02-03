package models

type Contact struct {
	Id          int    `json:"id"`
	UserId      int    `json:"user_id"`
	ContactId   int    `json:"contact_id"`
	ContactName string `json:"contact_name"`
}
