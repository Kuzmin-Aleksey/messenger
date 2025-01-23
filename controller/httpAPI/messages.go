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
	h.eventHandler.OnCreateMessage(m)
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

	h.eventHandler.OnUpdateMessage(m)
}

func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	var id id
	if err := json.NewDecoder(r.Body).Decode(&id); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}

	m, err := h.messages.DeleteMessage(r.Context(), id.Id)
	if err != nil {
		h.writeJSONError(w, err)
		return
	}

	h.eventHandler.OnDeleteMessage(id.Id, m.ChatId)
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

	messages, err := h.messages.GetFromChat(r.Context(), in.ChatId, in.LastId, in.Limit)
	if err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}

	h.writeJSON(w, http.StatusOK, messages)
}

func (h *Handler) GetMinMassageIdInChat(w http.ResponseWriter, r *http.Request) {

}
