package httpAPI

import (
	"encoding/json"
	"messanger/models"
	tr "messanger/pkg/error_trace"
	"net/http"
	"strconv"
)

func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var m *models.Message
	if err := json.NewDecoder(r.Body).Decode(m); err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("UserId").(int)
	exist, err := h.chats.CheckUserInChat(userId, m.ChatId)
	if err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !exist {
		h.info.Printf("user (%d) tried to create a message in the chat (%d)", userId, m.ChatId)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := h.messages.CreateMessage(m); err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.eventHandler.OnCreateMessage(m)
}

func (h *Handler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	var m *models.Message
	if err := json.NewDecoder(r.Body).Decode(m); err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("UserId").(int)
	exist, err := h.chats.CheckUserInChat(userId, m.ChatId)
	if err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !exist {
		h.info.Printf("user (%d) tried to delete a message in the chat (%d)", userId, m.ChatId)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := h.messages.UpdateMessage(m); err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.eventHandler.OnUpdateMessage(m)
}

func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	messageId, err := strconv.Atoi(r.Form.Get("message_id"))
	if err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	message, err := h.messages.GetById(messageId)
	if err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userId := r.Context().Value("UserId").(int)
	exist, err := h.chats.CheckUserInChat(userId, message.ChatId)
	if err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !exist {
		h.info.Printf("user (%d) tried to delete a message in the chat (%d)", userId, message.ChatId)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := h.messages.DeleteMessage(messageId); err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.eventHandler.OnDeleteMessage(messageId, message.ChatId)
}

type getMessagesInput struct {
	ChatId int `json:"chat_id"`
	LastId int `json:"last_id"`
	Limit  int `json:"limit"`
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	var in getMessagesInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("UserId").(int)
	exist, err := h.chats.CheckUserInChat(userId, in.ChatId)
	if err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !exist {
		h.info.Printf("user (%d) tried to delete a message in the chat (%d)", userId, in.ChatId)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	messages, err := h.messages.GetFromChat(in.ChatId, in.LastId, in.Limit)
	if err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(messages); err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
