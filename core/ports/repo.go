package ports

import (
	"messanger/domain"
	"messanger/pkg/errors"
)

type UsersRepo interface {
	New(user *domain.UserInfo) *errors.Error
	Update(user *domain.UserInfo) *errors.Error
	GetById(id int) (*domain.User, *errors.Error)
	GetInfoByEmail(email string) (*domain.UserInfo, *errors.Error)
	GetWithPass(email, password string) (*domain.UserInfo, *errors.Error)
	//IsExist(email string) (bool, *errors.Error)
	Delete(id int) *errors.Error
}

type ChatsRepo interface {
	New(chat *domain.Chat) *errors.Error
	Update(chat *domain.Chat) *errors.Error
	GetById(id int) (*domain.Chat, *errors.Error)
	AddUser(chatId int, userId int) *errors.Error
	CheckUserInChat(chatId int, userId int) (bool, *errors.Error)
	DeleteUser(chatId int, userId int) *errors.Error
	Delete(id int) *errors.Error
}

type MessagesRepo interface {
	New(message *domain.Message) *errors.Error
	GetByChat(chatId int, lastId int, count int) ([]domain.Message, *errors.Error)
	GetById(id int) (*domain.Message, *errors.Error)
	Update(id int, text string) *errors.Error
	Delete(id int) *errors.Error
	DeleteByChat(chatId int) *errors.Error
}
