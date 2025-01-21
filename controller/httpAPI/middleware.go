package httpAPI

import (
	"context"
	"messanger/pkg/errors"
	"net/http"
	"strings"
)

func (h *Handler) MwWithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		typeAndToken := strings.Split(r.Header.Get("Authorization"), " ")
		if len(typeAndToken) != 2 {
			h.writeJSONError(w, errors.New(r.Header.Get("Authorization"), "invalid token format", http.StatusUnauthorized))
			return
		}
		if typeAndToken[0] != "Bearer" {
			h.writeJSONError(w, errors.New(typeAndToken[0], "invalid auth type", http.StatusUnauthorized))
			return
		}
		token := typeAndToken[1]
		userId, err := h.auth.CheckAccessToken(token)
		if err != nil {
			h.writeJSONError(w, err)
			return
		}

		*r = *r.WithContext(context.WithValue(r.Context(), "UserId", userId))

		next(w, r)
	}
}
