package httpAPI

import (
	"messanger/core/service"
	"net/http"
)

type Logger interface {
	Println(v ...any)
	Printf(format string, v ...any)
}

type Handler struct {
	router *http.ServeMux

	auth     *service.AuthService
	users    *service.Users
	messages *service.Messages
	chats    *service.ChatService

	errors Logger
	info   Logger
}

func NewHandler() *Handler {
	return &Handler{
		router: http.NewServeMux(),
	}
}

func (h *Handler) InitRouter() {

}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}
