package chats

import (
	"context"
	"messanger/domain/models"
	"messanger/domain/ports"
	"messanger/domain/service/auth"
	"messanger/pkg/errors"
	"slices"
	"time"
)

type ChatService struct {
	chatsRepo  ports.ChatsRepo
	groupsRepo ports.GroupsRepo
}

func NewChatService(
	chatsRepo ports.ChatsRepo,
	groupsRepo ports.GroupsRepo,
) *ChatService {
	return &ChatService{
		chatsRepo:  chatsRepo,
		groupsRepo: groupsRepo,
	}
}

type ChatResponseUser struct {
	UserId int `json:"user_id"`
}

type UserRoleResponse struct {
	Id   int    `json:"id"`
	Role string `json:"role"`
}

type ChatResponseGroup struct {
	GroupId int                `json:"group_id"`
	Name    string             `json:"name"`
	Users   []UserRoleResponse `json:"users"`
}

type ChatResponse struct {
	ChatId          int        `json:"chat_id"`
	Type            string     `json:"type"`
	LastMessageTime *time.Time `json:"last_message_time,omitempty"`
	CreateTime      time.Time  `json:"create_time"`
	ChatInfo        any        `json:"chat_info"`
}

func (s *ChatService) GetAllUserChats(ctx context.Context) ([]*ChatResponse, *errors.Error) {
	userId := auth.ExtractUser(ctx)

	chats, err := s.chatsRepo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err.Trace()
	}

	resp := make([]*ChatResponse, 0, len(chats))
	for _, chat := range chats {
		chatResp := &ChatResponse{
			ChatId:     chat.Id,
			Type:       chat.Type,
			CreateTime: chat.CreateTime,
		}

		if chat.LastMessageTime.IsZero() {
			chatResp.LastMessageTime = nil
		} else {
			chatResp.LastMessageTime = &chat.LastMessageTime
		}

		switch chat.Type {
		case models.ChatTypeUser:
			userId, err := s.chatsRepo.GetUserCompanionByChatId(ctx, userId, chat.Id)
			if err != nil {
				return nil, err.Trace()
			}

			chatResp.ChatInfo = ChatResponseUser{
				UserId: userId,
			}

		case models.ChatTypeGroup:
			group, err := s.groupsRepo.GetGroupByChatId(ctx, chat.Id)
			if err != nil {
				return nil, err.Trace()
			}

			usersId, err := s.groupsRepo.GetUsersByGroup(ctx, group.Id)
			usersRole := make([]UserRoleResponse, 0, len(usersId))

			for _, userId := range usersId {
				role, err := s.groupsRepo.GetRole(ctx, userId, group.Id)
				if err != nil {
					return nil, err.Trace()
				}
				usersRole = append(usersRole, UserRoleResponse{
					Id:   userId,
					Role: role,
				})
			}

			chatResp.ChatInfo = ChatResponseGroup{
				GroupId: group.Id,
				Name:    group.Name,
				Users:   usersRole,
			}
		}

		resp = append(resp, chatResp)
	}

	slices.SortFunc(resp, func(chat1, chat2 *ChatResponse) int {
		var t1, t2 time.Time
		if chat1.LastMessageTime == nil {
			t1 = chat1.CreateTime
		} else {
			t1 = *chat1.LastMessageTime
		}
		if chat2.LastMessageTime == nil {
			t2 = chat2.CreateTime
		} else {
			t2 = *chat2.LastMessageTime
		}
		return int(t1.Sub(t2))
	})

	return resp, nil
}
