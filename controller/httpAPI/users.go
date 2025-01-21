package httpAPI

import (
	"encoding/json"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	info := new(domain.UserInfo)

	if err := json.NewDecoder(r.Body).Decode(info); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseJson, http.StatusBadRequest))
		return
	}

	if err := h.users.CreateUser(info); err != nil {
		h.writeJSONError(w, err)
		return
	}

	h.Login(w, r)
}
