package httpAPI

import (
	"encoding/json"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	user := new(domain.User)
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}

	if err := h.users.CreateUser(user); err != nil {
		h.writeJSONError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, newId(user.Id))
}

func (h *Handler) UpdateUserSelf(w http.ResponseWriter, r *http.Request) {
	user := new(domain.User)
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseForm, http.StatusBadRequest))
		return
	}

	lastPass := r.Form.Get("last_password") // to update password

	if err := h.users.UpdateUser(r.Context(), user, lastPass); err != nil {
		h.writeJSONError(w, err)
		return
	}
}

func (h *Handler) GetUserSelfInfo(w http.ResponseWriter, r *http.Request) {
	user, err := h.users.GetUserInfo(r.Context())
	if err != nil {
		h.writeJSONError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, user)
}

type addUserToChatRequest struct {
	UserId int    `json:"user_id"`
	ChatId int    `json:"chat_id"`
	Role   string `json:"role"`
}

func (h *Handler) AddUserToChat(w http.ResponseWriter, r *http.Request) {
	req := new(addUserToChatRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}
	if err := h.users.AddUserToChat(r.Context(), req.UserId, req.ChatId, req.Role); err != nil {
		h.writeJSONError(w, err)
		return
	}
}

func (h *Handler) GetUsersByChat(w http.ResponseWriter, r *http.Request) {
	chatId := new(id)
	if err := json.NewDecoder(r.Body).Decode(chatId); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}

	user, err := h.users.GetUsersByChat(r.Context(), chatId.Id)
	if err != nil {
		h.writeJSONError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, user)
}

type deleteUserToChatRequest struct {
	UserId int `json:"user_id"`
	ChatId int `json:"chat_id"`
}

func (h *Handler) DeleteUserFromChat(w http.ResponseWriter, r *http.Request) {
	req := new(deleteUserToChatRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}
	if err := h.users.DeleteUserFromChat(r.Context(), req.UserId, req.ChatId); err != nil {
		h.writeJSONError(w, err)
		return
	}
}

func (h *Handler) DeleteUserSelf(w http.ResponseWriter, r *http.Request) {
	if err := h.users.DeleteUser(r.Context()); err != nil {
		h.writeJSONError(w, err)
		return
	}
}
