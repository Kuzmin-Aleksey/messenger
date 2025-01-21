package httpAPI

import (
	"encoding/json"
	"fmt"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
	"strconv"
)

func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	m := new(domain.Message)
	if err := json.NewDecoder(r.Body).Decode(m); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}

	userId := r.Context().Value("UserId").(int)
	exist, err := h.chats.CheckUserInChat(userId, m.ChatId)
	if err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}

	if !exist {
		h.writeJSONError(w, errors.New(fmt.Sprintf("user (%d) tried to create a message in the chat (%d)", userId, m.ChatId),
			"forbidden", http.StatusForbidden))
		return
	}

	if err := h.messages.CreateMessage(m); err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}

	h.eventHandler.OnCreateMessage(m)
}

func (h *Handler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	m := new(domain.Message)
	if err := json.NewDecoder(r.Body).Decode(m); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}

	userId := r.Context().Value("UserId").(int)
	exist, err := h.chats.CheckUserInChat(userId, m.ChatId)
	if err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}
	if !exist {
		h.writeJSONError(w, errors.New(fmt.Sprintf("user (%d) tried to delete a message in the chat (%d)", userId, m.ChatId),
			"forbidden", http.StatusForbidden))
		return
	}

	if err := h.messages.UpdateMessage(m); err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}

	h.eventHandler.OnUpdateMessage(m)
}

func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	if e := r.ParseForm(); e != nil {
		h.writeJSONError(w, errors.New(e, domain.ErrParseForm, http.StatusBadRequest))
		return
	}
	messageId, e := strconv.Atoi(r.Form.Get("message_id"))
	if e != nil {
		h.writeJSONError(w, errors.New(e, "invalid message_id", http.StatusBadRequest))
		return
	}
	m, err := h.messages.GetById(messageId)
	if err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}

	userId := r.Context().Value("UserId").(int)
	exist, err := h.chats.CheckUserInChat(userId, m.ChatId)
	if err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}
	if !exist {
		h.writeJSONError(w, errors.New(fmt.Sprintf("user (%d) tried to delete a message in the chat (%d)", userId, m.ChatId),
			"forbidden", http.StatusForbidden))
		return
	}

	if err := h.messages.DeleteMessage(messageId); err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}

	h.eventHandler.OnDeleteMessage(messageId, m.ChatId)
}

type getMessagesInput struct {
	ChatId int `json:"chat_id"`
	LastId int `json:"last_id"`
	Limit  int `json:"limit"`
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	var in getMessagesInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}

	userId := r.Context().Value("UserId").(int)
	exist, err := h.chats.CheckUserInChat(userId, in.ChatId)
	if err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}
	if !exist {
		h.writeJSONError(w, errors.New(fmt.Sprintf("user (%d) tried to delete a message in the chat (%d)", userId, in.ChatId),
			"forbidden", http.StatusForbidden))
		return
	}

	messages, err := h.messages.GetFromChat(in.ChatId, in.LastId, in.Limit)
	if err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}

	h.writeJSON(w, http.StatusOK, messages)
}
