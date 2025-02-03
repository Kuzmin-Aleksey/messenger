package http

import (
	"messanger/domain/service/auth"
	"messanger/pkg/errors"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) MwWithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

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
		userId, err := h.auth.DecodeAccessToken(token)
		if err != nil {
			h.writeJSONError(w, err)
			return
		}

		//user, _ := h.users.GetById(ctx, userId)

		r = r.WithContext(auth.CtxWithUser(ctx, userId))

		next(w, r)
	}
}

type CustomWriter struct {
	http.ResponseWriter
	http.Hijacker
	StatusCode    int
	ContentLength int
}

func (r *CustomWriter) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *CustomWriter) Write(p []byte) (int, error) {
	n, err := r.ResponseWriter.Write(p)
	r.ContentLength += n
	return n, err
}

func (h *Handler) MwLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		customWriter := &CustomWriter{
			ResponseWriter: w,
			Hijacker:       w.(http.Hijacker),
			StatusCode:     http.StatusOK,
		}

		start := time.Now()
		next(customWriter, r)
		end := time.Now()

		h.info.Log(r.Method, r.URL.Path, customWriter.StatusCode, r.RemoteAddr, uint64(r.ContentLength), uint64(customWriter.ContentLength), end.Sub(start))
	}
}
