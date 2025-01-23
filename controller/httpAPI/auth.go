package httpAPI

import (
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseForm, http.StatusBadRequest))
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	if email == "" || password == "" {
		h.writeJSONError(w, errors.New(email+":"+password, "missing form values (email,password)", http.StatusBadRequest))
		return
	}

	tokens, err := h.auth.Login(email, password)
	if err != nil {
		h.writeJSONError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, tokens)
}

func (h *Handler) UpdateTokens(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseForm, http.StatusBadRequest))
		return
	}

	refreshToken := r.Form.Get("refresh_token")
	if refreshToken == "" {
		h.writeJSONError(w, errors.New(refreshToken, "missing refresh token", http.StatusBadRequest))
		return
	}

	tokens, err := h.auth.UpdateTokens(refreshToken)
	if err != nil {
		h.writeJSONError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, tokens)
}
