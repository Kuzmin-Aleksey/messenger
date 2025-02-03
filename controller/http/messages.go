package http

import (
	"encoding/json"
	"messanger/domain/models"
	"messanger/domain/service/messages"
	"messanger/pkg/errors"
	"net/http"
	"strconv"
)

func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	m := new(models.Message)
	if err := json.NewDecoder(r.Body).Decode(m); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseJson, http.StatusBadRequest))
		return
	}
	if err := h.messages.CreateMessage(r.Context(), m); err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}
}

func (h *Handler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseForm, http.StatusBadRequest))
		return
	}
	messageId, _ := strconv.Atoi(r.Form.Get("message_id"))

	dto := new(messages.UpdateMessageDTO)
	if err := json.NewDecoder(r.Body).Decode(dto); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseJson, http.StatusBadRequest))
		return
	}

	if err := h.messages.UpdateMessage(r.Context(), messageId, dto); err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}

}

func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseForm, http.StatusBadRequest))
		return
	}
	messageId, _ := strconv.Atoi(r.Form.Get("message_id"))

	if err := h.messages.DeleteMessage(r.Context(), messageId); err != nil {
		h.writeJSONError(w, err)
		return
	}
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	dto := new(messages.GetMessagesDTO)
	if err := json.NewDecoder(r.Body).Decode(dto); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseJson, http.StatusBadRequest))
		return
	}

	resp, err := h.messages.GetFromChat(r.Context(), dto)
	if err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

type GetMinMassageIdInChatResponse struct {
	Id int `json:"id"`
}

// GetMinMassageIdInChat to know where end of chat
func (h *Handler) GetMinMassageIdInChat(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseForm, http.StatusBadRequest))
		return
	}
	chatId, _ := strconv.Atoi(r.Form.Get("chat_id"))

	messageId, err := h.messages.GetMinMassageIdInChat(r.Context(), chatId)
	if err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}

	h.writeJSON(w, http.StatusOK, GetMinMassageIdInChatResponse{
		Id: messageId,
	})
}
