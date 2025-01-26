package httpAPI

import (
	"encoding/json"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
)

func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	m := new(domain.Message)
	if err := json.NewDecoder(r.Body).Decode(m); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}
	if err := h.messages.CreateMessage(r.Context(), m); err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}
	go h.eventHandler.OnCreateMessage(m)
}

func (h *Handler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	m := new(domain.Message)
	if err := json.NewDecoder(r.Body).Decode(m); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}

	if err := h.messages.UpdateMessage(r.Context(), m); err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}

	go h.eventHandler.OnUpdateMessage(m)
}

func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	messageId := new(id)
	if err := json.NewDecoder(r.Body).Decode(messageId); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}

	m, err := h.messages.DeleteMessage(r.Context(), messageId.Id)
	if err != nil {
		h.writeJSONError(w, err)
		return
	}

	go h.eventHandler.OnDeleteMessage(messageId.Id, m.ChatId)
}

type getMessagesInput struct {
	ChatId int `json:"chat_id"`
	LastId int `json:"last_id"`
	Count  int `json:"count"`
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	var in getMessagesInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}

	messages, err := h.messages.GetFromChat(r.Context(), in.ChatId, in.LastId, in.Count)
	if err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}

	h.writeJSON(w, http.StatusOK, messages)
}

// GetMinMassageIdInChat to know where end of chat
func (h *Handler) GetMinMassageIdInChat(w http.ResponseWriter, r *http.Request) {
	chatId := new(id)
	if err := json.NewDecoder(r.Body).Decode(&chatId); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}

	messageId, err := h.messages.GetMinMassageIdInChat(chatId.Id)
	if err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}

	h.writeJSON(w, http.StatusOK, newId(messageId))
}
