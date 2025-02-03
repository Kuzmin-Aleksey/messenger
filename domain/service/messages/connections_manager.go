package messages

import (
	"context"
	"messanger/domain/models"
	"messanger/pkg/errors"
	"sync"
	"time"
)

type ChatsGetter interface {
	GetChatListByUser(ctx context.Context, userId int) ([]int, *errors.Error)
}

type Conn interface {
	Send(*Event) (ok bool)
	Ping() (ok bool)
}

type UserConnections struct {
	connections *[]Conn
	listenChats []int
}

const (
	EventTypeCreate  = "create"
	EventTypeUpdate  = "update"
	EventTypeDelete  = "delete"
	EventTypeOnline  = "user_online"
	EventTypeOffline = "user_offline"
)

type Event struct {
	Type   string `json:"type"`
	ChatId int    `json:"chat_id"`
	Data   any    `json:"data"`
}

type ConnectionsManager struct {
	chatToUsers map[int]map[int]*[]Conn
	userChats   map[int]UserConnections
	chatsGetter ChatsGetter
	mu          sync.RWMutex
}

func (s *MessagesService) NewConnectionsManager() *ConnectionsManager {
	if s.connManager != nil {
		return s.connManager
	}
	m := &ConnectionsManager{
		chatToUsers: make(map[int]map[int]*[]Conn),
		userChats:   make(map[int]UserConnections),
		chatsGetter: s.userChats,
	}
	s.connManager = m

	go func() {
		for {
			time.Sleep(5 * time.Second)
			for userId, connections := range m.userChats {
				for i, conn := range *connections.connections {
					if !conn.Ping() {
						m.removeConn(userId, i)
					}
				}
			}
		}
	}()
	return m
}

func (m *ConnectionsManager) InsertConn(ctx context.Context, userId int, conn Conn) *errors.Error {
	if m.CheckOnline(userId) {
		m.insertConn(userId, conn)
		return nil
	}
	userChats, err := m.chatsGetter.GetChatListByUser(ctx, userId)
	if err != nil {
		return err.Trace()
	}
	m.insertConn(userId, conn, userChats...)
	return nil
}

func (m *ConnectionsManager) insertConn(userId int, conn Conn, chats ...int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if chats != nil {
		connections := &[]Conn{conn}
		m.userChats[userId] = UserConnections{
			listenChats: chats,
			connections: connections,
		}
		for _, chatId := range chats {
			if _, ok := m.chatToUsers[chatId]; !ok {
				m.chatToUsers[chatId] = make(map[int]*[]Conn)
			}
			m.chatToUsers[chatId][userId] = connections
		}
	} else {
		*m.userChats[userId].connections = append(*m.userChats[userId].connections, conn)
	}
}

func (m *ConnectionsManager) removeConn(userId int, i int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(*m.userChats[userId].connections) != 1 {
		*m.userChats[userId].connections = append((*m.userChats[userId].connections)[:i], (*m.userChats[userId].connections)[i+1:]...)
		return
	}
	chats := m.userChats[userId].listenChats
	delete(m.userChats, userId)

	for _, chatId := range chats {
		delete(m.chatToUsers[chatId], userId)
		if len(m.chatToUsers[chatId]) == 0 {
			delete(m.chatToUsers, chatId)
		}
	}
	delete(m.userChats, userId)

	// todo - update last online time
}

func (m *ConnectionsManager) CheckOnlineList(usersId []int) []bool {
	res := make([]bool, len(usersId))
	for i, userId := range usersId {
		res[i] = m.CheckOnline(userId)
	}
	return res
}

func (m *ConnectionsManager) CheckOnline(userId int) bool {
	_, ok := m.userChats[userId]
	return ok
}

func (m *ConnectionsManager) sendEventToChat(chatId int, event *Event) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	event.ChatId = chatId

	for userId, connections := range m.chatToUsers[chatId] {
		for i, conn := range *connections {
			if !conn.Send(event) {
				m.removeConn(userId, i)
			}
		}
	}
}

func (m *ConnectionsManager) onCreateMessage(msg *models.Message) {
	m.sendEventToChat(msg.ChatId, &Event{
		Type: EventTypeCreate,
		Data: msg,
	})
}

func (m *ConnectionsManager) onUpdateMessage(msg *models.Message) {
	m.sendEventToChat(msg.ChatId, &Event{
		Type: EventTypeUpdate,
		Data: msg,
	})
}

func (m *ConnectionsManager) onDeleteMessage(id int, chatId int) {
	m.sendEventToChat(chatId, &Event{
		Type: EventTypeDelete,
		Data: struct {
			Id int `json:"id"`
		}{
			Id: id,
		},
	})
}
