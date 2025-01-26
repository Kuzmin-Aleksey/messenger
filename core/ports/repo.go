package ports

import (
	"messanger/domain"
	"messanger/pkg/errors"
)

type UsersRepo interface {
	New(user *domain.User) *errors.Error
	SetConfirm(userId int, v bool) *errors.Error
	AddUserToChat(userId int, chatId int, role string) *errors.Error
	CheckUserInChat(userId int, chatId int) (bool, *errors.Error)
	DeleteUserFromChat(chatId int, userId int) *errors.Error
	SetRole(userId int, chatId int, role string) *errors.Error
	GetRole(userId int, chatId int) (string, *errors.Error)
	GetUsersByChat(chatId int) ([]domain.User, *errors.Error)
	CountOfUsersInChat(chatId int) (int, *errors.Error)
	Update(user *domain.User) *errors.Error
	GetById(id int) (*domain.User, *errors.Error)
	GetByEmail(email string) (*domain.User, *errors.Error)
	GetByEmailWithPass(email, password string) (*domain.User, *errors.Error)
	CheckPass(id int, password string) (bool, *errors.Error)
	Delete(id int) *errors.Error
}

type ChatsRepo interface {
	New(chat *domain.Chat, creator int) *errors.Error
	Update(chat *domain.Chat) *errors.Error
	GetChatsByUser(userId int) ([]domain.Chat, *errors.Error)
	GetById(id int) (*domain.Chat, *errors.Error)
	GetUserRole(userId int, chatId int) (string, *errors.Error)
	Delete(id int) *errors.Error
}

type MessagesRepo interface {
	New(message *domain.Message) *errors.Error
	GetByChat(chatId int, lastId int, count int) ([]domain.Message, *errors.Error)
	GetMinMassageIdInChat(chatId int) (int, *errors.Error)
	GetById(id int) (*domain.Message, *errors.Error)
	Update(id int, text string) *errors.Error
	Delete(id int) *errors.Error
	DeleteByChat(chatId int) *errors.Error
}
