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
	users    *service.UsersService
	messages *service.MessagesService
	chats    *service.ChatService

	errors Logger
	info   Logger

	eventHandler *EventHandler
}

func NewHandler(
	auth *service.AuthService,
	users *service.UsersService,
	messages *service.MessagesService,
	chats *service.ChatService,

	errors Logger,
	info Logger,
) *Handler {
	return &Handler{
		auth:         auth,
		users:        users,
		messages:     messages,
		chats:        chats,
		errors:       errors,
		info:         info,
		router:       http.NewServeMux(),
		eventHandler: NewEventHandler(),
	}
}

func (h *Handler) InitRouter() {
	h.router.Handle("/register", h)

}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}
