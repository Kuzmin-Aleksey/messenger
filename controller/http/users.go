package http

import (
	"encoding/json"
	"messanger/domain/models"
	"messanger/domain/service/users"
	"messanger/pkg/errors"
	"net/http"
	"strconv"
	"time"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseForm, http.StatusBadRequest))
		return
	}

	user := &users.CreateUserDTO{
		Phone:    r.Form.Get("phone"),
		Password: r.Form.Get("password"),
		Name:     r.Form.Get("name"),
		RealName: r.Form.Get("real_name"),
	}

	if err := h.users.CreateUser(r.Context(), user); err != nil {
		h.writeJSONError(w, err)
		return
	}
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseForm, http.StatusBadRequest))
		return
	}

	if err := h.users.UpdateUser(r.Context(), &users.UpdateUserDTO{
		Phone:       r.Form.Get("phone"),
		OldPassword: r.Form.Get("old_password"),
		Password:    r.Form.Get("new_password"),
		Name:        r.Form.Get("username"),
		RealName:    r.Form.Get("real_name"),
	}); err != nil {
		h.writeJSONError(w, err)
		return
	}
}

func (h *Handler) FindUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseForm, http.StatusBadRequest))
		return
	}
	var params users.FindUserDTO
	params.UserId, _ = strconv.Atoi(r.Form.Get("id"))
	params.Username = r.Form.Get("username")
	params.Phone = r.Form.Get("phone")

	user, err := h.users.FindUser(r.Context(), &params)
	if err != nil {
		h.writeJSONError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, user)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if err := h.users.DeleteUser(r.Context()); err != nil {
		h.writeJSONError(w, err)
		return
	}
}

func (h *Handler) CreateChatWithUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseForm, http.StatusBadRequest))
		return
	}
	userId, e := strconv.Atoi(r.Form.Get("user_id"))
	if e != nil {
		h.writeJSONError(w, errors.New(e, "invalid user id", http.StatusBadRequest))
	}

	chat, err := h.users.CreateChatWithUser(r.Context(), userId)
	if err != nil {
		h.writeJSONError(w, err)
	}
	h.writeJSON(w, http.StatusOK, chat)
}

func (h *Handler) AddContact(w http.ResponseWriter, r *http.Request) {
	req := new(users.CreateContactDTO)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseJson, http.StatusBadRequest))
		return
	}
	if err := h.users.AddToContact(r.Context(), req); err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}
}

func (h *Handler) GetUserContacts(w http.ResponseWriter, r *http.Request) {
	contacts, err := h.users.GetUserContacts(r.Context())
	if err != nil {
		h.writeJSONError(w, err)
	}
	h.writeJSON(w, http.StatusOK, contacts)
}

type removeContactRequest struct {
	ContactUserId int    `json:"contact_user_id"`
	Name          string `json:"name"`
}

func (h *Handler) RenameContact(w http.ResponseWriter, r *http.Request) {
	req := new(removeContactRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseJson, http.StatusBadRequest))
		return
	}
	if err := h.users.RenameContact(r.Context(), req.ContactUserId, req.Name); err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}
}

func (h *Handler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseForm, http.StatusBadRequest))
		return
	}
	userId, _ := strconv.Atoi(r.Form.Get("contact_user_id"))

	if err := h.users.RemoveUserFromContact(r.Context(), userId); err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}
}

type CheckOnlineRequest struct {
	UsersId []int `json:"users_id"`
}

type CheckOnlineResponse struct {
	UserId   int        `json:"user_id"`
	Online   bool       `json:"online"`
	LastSeen *time.Time `json:"last_seen,omitempty"`
}

func (h *Handler) CheckOnline(w http.ResponseWriter, r *http.Request) {
	req := new(CheckOnlineRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseJson, http.StatusBadRequest))
		return
	}
	resp := make([]CheckOnlineResponse, len(req.UsersId))

	statuses := h.connManager.CheckOnlineList(req.UsersId)

	var err *errors.Error
	for i, status := range statuses {
		var lastSeen *time.Time
		if !status {
			lastSeen, err = h.users.GetLastOnline(r.Context(), req.UsersId[i])
			if err != nil {
				h.writeJSONError(w, err)
				return
			}
		}
		resp[i] = CheckOnlineResponse{
			UserId:   req.UsersId[i],
			Online:   status,
			LastSeen: lastSeen,
		}
	}

	h.writeJSON(w, http.StatusOK, resp)
}
