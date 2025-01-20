package ports

import "messanger/models"

type UsersRepo interface {
	New(user *models.User) error
	GetById(id int) (*models.User, error)
	AddChat(userId int, chatId int) error
	Delete(id int) error
}

type ChatsRepo interface {
	New(chat *models.Chat) error
	GetById(id int) (*models.Chat, error)
	AddUser(chatId int, userId int) error
	Delete(id int) error
}

type MessagesRepo interface {
	New(message *models.Message) error
	GetByChat(chatId int, offset int, count int) ([]models.Message, error)
	Delete(id int) error
}

type AuthUsers interface {
	New(user *models.AuthUser) error
	GetUserByEmail(email string) (*models.AuthUser, error)
	CheckPassword(password string, email string) (bool, error)
	IsExist(email string) (bool, error)
	Delete(Id int) error
}
