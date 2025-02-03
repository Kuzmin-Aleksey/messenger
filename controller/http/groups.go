package http

import (
	"encoding/json"
	"messanger/domain/models"
	"messanger/domain/service/groups"
	"messanger/pkg/errors"
	"net/http"
	"strconv"
)

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	group := new(models.Group)
	if err := json.NewDecoder(r.Body).Decode(group); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseJson, http.StatusBadRequest))
		return
	}
	if err := h.groups.NewGroup(r.Context(), group); err != nil {
		h.writeJSONError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, group)
}

func (h *Handler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseForm, http.StatusBadRequest))
		return
	}
	groupId, _ := strconv.Atoi(r.Form.Get("group_id"))

	dto := new(groups.UpdateGroupDTO)
	if err := json.NewDecoder(r.Body).Decode(dto); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseJson, http.StatusBadRequest))
		return
	}

	if err := h.groups.UpdateGroup(r.Context(), groupId, dto); err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}
}

type userIdAndGroupIdRequest struct {
	GroupId int `json:"group_id"`
	UserId  int `json:"user_id"`
}

func (h *Handler) AddUserToGroup(w http.ResponseWriter, r *http.Request) {
	req := new(userIdAndGroupIdRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseJson, http.StatusBadRequest))
		return
	}
	if err := h.groups.AddUserToGroup(r.Context(), req.GroupId, req.UserId); err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}
}

func (h *Handler) DeleteUserFromGroup(w http.ResponseWriter, r *http.Request) {
	req := new(userIdAndGroupIdRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseJson, http.StatusBadRequest))
		return
	}
	if err := h.groups.RemoveUserFromGroup(r.Context(), req.GroupId, req.UserId); err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}
}

type setUserRoleInGroupRequest struct {
	Role string `json:"role"`
}

func (h *Handler) SetUsersRoleInGroup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseForm, http.StatusBadRequest))
		return
	}
	req := new(setUserRoleInGroupRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseJson, http.StatusBadRequest))
	}

	groupId, _ := strconv.Atoi(r.Form.Get("group_id"))
	userId, _ := strconv.Atoi(r.Form.Get("user_id"))

	if err := h.groups.SetUsersRole(r.Context(), groupId, userId, req.Role); err != nil {
		h.writeJSONError(w, err)
		return
	}
}

func (h *Handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseForm, http.StatusBadRequest))
		return
	}
	groupId, _ := strconv.Atoi(r.Form.Get("group_id"))

	if err := h.groups.RemoveGroup(r.Context(), groupId); err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}
}

func (h *Handler) GetUsersByGroup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseForm, http.StatusBadRequest))
		return
	}
	groupId, _ := strconv.Atoi(r.Form.Get("group_id"))

	u, err := h.groups.GetUsersByGroup(r.Context(), groupId)
	if err != nil {
		h.writeJSONError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, u)
}
