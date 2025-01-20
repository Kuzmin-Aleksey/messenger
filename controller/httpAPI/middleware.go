package httpAPI

import (
	"context"
	tr "messanger/pkg/error_trace"
	"net/http"
	"strings"
)

func (h *Handler) MwWithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		typeAndToken := strings.Split(r.Header.Get("Authorization"), " ")
		if len(typeAndToken) != 2 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if typeAndToken[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		token := typeAndToken[1]
		userId, err := h.auth.CheckAccessToken(token)
		if err != nil {
			h.errors.Println(tr.Trace(err))
			w.WriteHeader(http.StatusUnauthorized)
		}

		*r = *r.WithContext(context.WithValue(r.Context(), "UserId", userId))

		next(w, r)
	}
}
