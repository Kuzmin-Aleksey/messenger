package httpAPI

import (
	"encoding/json"
	tr "messanger/pkg/error_trace"
	"net/http"
)

func (h *Handler) AuthLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	if email == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokens, err := h.auth.Login(email, password)
	if err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := json.NewEncoder(w).Encode(tokens); err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
	h.info.Printf("User %s logged in", email)
}

func (h *Handler) AuthUpdateTokens(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	refreshToken := r.Form.Get("refresh_token")
	if refreshToken == "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	tokens, err := h.auth.UpdateTokens(refreshToken)
	if err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := json.NewEncoder(w).Encode(tokens); err != nil {
		h.errors.Println(tr.Trace(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
