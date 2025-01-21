package httpAPI

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
	"strconv"
)

const (
	EventTypeCreate = "create"
	EventTypeUpdate = "update"
	EventTypeDelete = "delete"
)

type Event struct {
	Type string
	Data any
}

type Channel struct {
	Channel       chan *Event
	CountOfListen int
}

type EventHandler struct {
	connects   map[int]map[int]*websocket.Conn
	wsUpgrader *websocket.Upgrader
	errors     Logger
}

func NewEventHandler() *EventHandler {
	h := &EventHandler{
		connects: make(map[int]map[int]*websocket.Conn),
		wsUpgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	return h
}

func (e *EventHandler) OnCreateMessage(m *domain.Message) {
	e.writeToChat(m.ChatId, &Event{
		Type: EventTypeCreate,
		Data: m,
	})
}

func (e *EventHandler) OnUpdateMessage(m *domain.Message) {
	e.writeToChat(m.ChatId, &Event{
		Type: EventTypeUpdate,
		Data: m,
	})
}

func (e *EventHandler) OnDeleteMessage(id int, chatId int) {
	e.writeToChat(chatId, &Event{
		Type: EventTypeDelete,
		Data: id,
	})
}

func (e *EventHandler) writeToChat(chatId int, v any) {
	chat := e.connects[chatId]
	if chat == nil {
		return
	}

	for userId, conn := range chat {
		if err := conn.WriteJSON(v); err != nil {
			e.errors.Println(fmt.Errorf("error on WriteJSON: %v, user id: %d", errors.Trace(err), userId))
			conn.Close()
			e.removeConnect(chatId, userId)
		}
	}
}

func (e *EventHandler) addConnect(chatId int, userId int, c *websocket.Conn) {
	if _, ok := e.connects[chatId]; !ok {
		e.connects[chatId] = make(map[int]*websocket.Conn)
	}
	e.connects[chatId][userId] = c
}

func (e *EventHandler) removeConnect(chatId int, userId int) {
	delete(e.connects[chatId], userId)
	if len(e.connects[chatId]) == 0 {
		delete(e.connects, chatId)
	}
}

func (e *EventHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		e.errors.Println(errors.Trace(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(responseError{domain.ErrParseForm})
		return
	}

	chatId, err := strconv.Atoi(r.Form.Get("chat_id"))
	if err != nil {
		e.errors.Println(errors.Trace(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(responseError{"invalid chat_id parameter"})
		return
	}

	ws, err := e.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		e.errors.Println(errors.Trace(err))
		return
	}

	e.addConnect(chatId, r.Context().Value("UserId").(int), ws)
	ws.ReadMessage()
}
