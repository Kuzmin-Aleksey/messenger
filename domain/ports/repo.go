package ports

import (
	"context"
	"messanger/domain/models"
	"messanger/pkg/errors"
	"time"
)

type UsersRepo interface {
	New(ctx context.Context, user *models.User) *errors.Error
	SetConfirm(ctx context.Context, userId int, v bool) *errors.Error
	UpdateUsername(ctx context.Context, userId int, name string) *errors.Error
	UpdateRealName(ctx context.Context, userId int, realName string) *errors.Error
	UpdatePassword(ctx context.Context, userId int, password string) *errors.Error
	UpdatePhone(ctx context.Context, userId int, phone string) *errors.Error
	UpdateLastOnlineTime(ctx context.Context, userId int, time time.Time) *errors.Error
	GetLastOnline(ctx context.Context, userId int) (time.Time, *errors.Error)
	GetById(ctx context.Context, id int) (*models.User, *errors.Error)
	FindByPhone(ctx context.Context, phone string) (*models.User, *errors.Error)
	FindByName(ctx context.Context, name string) (*models.User, *errors.Error)
	GetByPhoneWithPass(ctx context.Context, phone, password string) (*models.User, *errors.Error)
	GetByIdWithPass(ctx context.Context, id int, password string) (*models.User, *errors.Error)
	Delete(ctx context.Context, id int) *errors.Error
}

type ContactsRepo interface {
	Create(ctx context.Context, contact *models.Contact) *errors.Error
	SetContactName(ctx context.Context, userid, contactId int, name string) *errors.Error
	GetContactsByUser(ctx context.Context, userId int) ([]models.Contact, *errors.Error)
	GetContact(ctx context.Context, userId, contactId int) (*models.Contact, *errors.Error)
	Delete(ctx context.Context, userId, contactId int) (e *errors.Error)
}

type GroupsRepo interface {
	New(ctx context.Context, group *models.Group) *errors.Error
	Update(ctx context.Context, group *models.Group) *errors.Error
	GetGroupsByUser(ctx context.Context, userId int) ([]models.Group, *errors.Error)
	GetGroupByChatId(ctx context.Context, chatId int) (*models.Group, *errors.Error)
	GetUsersByGroup(ctx context.Context, id int) ([]int, *errors.Error)
	GetById(ctx context.Context, id int) (*models.Group, *errors.Error)
	SetRole(ctx context.Context, userId int, groupId int, role string) *errors.Error
	GetRole(ctx context.Context, userId int, groupId int) (string, *errors.Error)
	Delete(ctx context.Context, id int) (e *errors.Error)
}

type ChatsRepo interface {
	New(ctx context.Context, chat *models.Chat) *errors.Error
	UpdateTime(ctx context.Context, chatId int, time time.Time) *errors.Error
	AddUserToChat(ctx context.Context, id int, userId int) *errors.Error
	RemoveUserFromChat(ctx context.Context, id int, userId int) *errors.Error
	CheckUserInChat(ctx context.Context, userId int, chatId int) (bool, *errors.Error)
	CountUsersInChat(ctx context.Context, id int) (int, *errors.Error)
	GetByUserId(ctx context.Context, userId int) ([]models.Chat, *errors.Error)
	GetChatListByUser(ctx context.Context, userId int) ([]int, *errors.Error)
	GetUserCompanionByChatId(ctx context.Context, userId int, chatId int) (int, *errors.Error)
	GetById(ctx context.Context, id int) (*models.Chat, *errors.Error)
	Delete(ctx context.Context, id int) *errors.Error
}

type MessagesRepo interface {
	New(ctx context.Context, message *models.Message) *errors.Error
	GetByChat(ctx context.Context, chatId int, lastId int, count int) ([]models.Message, *errors.Error)
	GetMinMassageIdInChat(ctx context.Context, chatId int) (int, *errors.Error)
	IsUserMessage(ctx context.Context, id int, userId int) (bool, *errors.Error)
	GetById(ctx context.Context, id int) (*models.Message, *errors.Error)
	Update(ctx context.Context, id int, text string) *errors.Error
	Delete(ctx context.Context, id int) *errors.Error
}
