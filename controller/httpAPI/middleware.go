package httpAPI

import (
	"context"
	"messanger/pkg/errors"
	"net/http"
	"strings"
	"time"
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
		userId, err := h.auth.TokenManager.DecodeToken(token)
		if err != nil {
			h.writeJSONError(w, err)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "UserId", userId))

		next(w, r)
	}
}

type CustomWriter struct {
	http.ResponseWriter
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
			StatusCode:     http.StatusOK,
		}

		start := time.Now()
		next(customWriter, r)
		end := time.Now()

		h.info.Log(r.Method, r.URL.Path, customWriter.StatusCode, r.RemoteAddr, uint64(r.ContentLength), uint64(customWriter.ContentLength), end.Sub(start))
	}
}

type WriterHijacker struct {
	http.ResponseWriter
	http.Hijacker
}

func (h *Handler) MwWithHijacker(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(http.Hijacker); ok {
			next(w, r)
		}
		switch v := w.(type) {
		case http.Hijacker:
			next(w, r)
		case *CustomWriter:
			wHijacker := WriterHijacker{
				ResponseWriter: w,
				Hijacker:       v.ResponseWriter.(http.Hijacker),
			}
			next(wHijacker, r)
		default:
			h.errors.Printf("not found hijacker in writer")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
