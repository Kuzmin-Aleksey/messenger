package httpAPI

import (
	"encoding/json"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
)

func (h *Handler) CreateChat(w http.ResponseWriter, r *http.Request) {
	chat := new(domain.Chat)
	if err := json.NewDecoder(r.Body).Decode(chat); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}

	if err := h.chats.NewChat(r.Context(), chat); err != nil {
		h.writeJSONError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, newId(chat.Id))
}

func (h *Handler) UpdateChat(w http.ResponseWriter, r *http.Request) {
	chat := new(domain.Chat)
	if err := json.NewDecoder(r.Body).Decode(chat); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}
	if err := h.chats.UpdateChat(r.Context(), chat); err != nil {
		h.writeJSONError(w, err)
		return
	}
}

func (h *Handler) GetUserChats(w http.ResponseWriter, r *http.Request) {
	chats, err := h.chats.GetUserChats(r.Context())
	if err != nil {
		h.writeJSONError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, chats)
}

func (h *Handler) DeleteChat(w http.ResponseWriter, r *http.Request) {
	chatId := new(id)
	if err := json.NewDecoder(r.Body).Decode(chatId); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}
	if err := h.chats.Delete(r.Context(), chatId.Id); err != nil {
		h.writeJSONError(w, err)
		return
	}
}
