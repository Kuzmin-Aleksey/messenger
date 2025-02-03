package http

import (
	"messanger/domain/models"
	"messanger/pkg/errors"
	"net/http"
)

func (h *Handler) Login1FA(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseForm, http.StatusBadRequest))
		return
	}

	phone := r.Form.Get("phone")
	password := r.Form.Get("password")

	err := h.auth.Login1FA(r.Context(), phone, password)
	if err != nil {
		h.writeJSONError(w, err)
		return
	}
}

func (h *Handler) Login2FA(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseForm, http.StatusBadRequest))
		return
	}

	phone := r.Form.Get("phone")
	code := r.Form.Get("code")

	tokens, err := h.auth.Login2FA(r.Context(), phone, code)
	if err != nil {
		h.writeJSONError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, tokens)
}

func (h *Handler) UpdateTokens(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, models.ErrParseForm, http.StatusBadRequest))
		return
	}
	refreshToken := r.Form.Get("refresh_token")

	tokens, err := h.auth.UpdateTokens(r.Context(), refreshToken)
	if err != nil {
		h.writeJSONError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, tokens)
}
