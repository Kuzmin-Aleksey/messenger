package httpAPI

import (
	"github.com/gorilla/mux"
	"messanger/core/service"
	"net/http"
)

type Logger interface {
	Println(v ...any)
	Printf(format string, v ...any)
}

type Handler struct {
	router *mux.Router

	auth     *service.AuthService
	users    *service.UsersService
	messages *service.MessagesService
	chats    *service.ChatService

	errors Logger
	info   *HttpLogger

	eventHandler *EventHandler
}

func NewHandler(
	auth *service.AuthService,
	users *service.UsersService,
	messages *service.MessagesService,
	chats *service.ChatService,

	errors Logger,
	// info io.Writer,
) *Handler {
	return &Handler{
		auth:         auth,
		users:        users,
		messages:     messages,
		chats:        chats,
		errors:       errors,
		info:         NewHttpLogger(),
		router:       mux.NewRouter(),
		eventHandler: NewEventHandler(),
	}
}

func (h *Handler) InitRouter() {
	h.router.HandleFunc("/auth/register", h.MwLogging(h.Register)).Methods(http.MethodPost)
	h.router.HandleFunc("/auth/login", h.MwLogging(h.Register)).Methods(http.MethodPost)
	h.router.HandleFunc("/auth/refresh-tokens", h.MwLogging(h.UpdateTokens)).Methods(http.MethodPost)

	h.router.HandleFunc("/self/update", h.MwWithAuth(h.MwLogging(h.UpdateUserSelf))).Methods(http.MethodPost)
	h.router.HandleFunc("/self/delete", h.MwWithAuth(h.MwLogging(h.DeleteUserSelf))).Methods(http.MethodPost)
	h.router.HandleFunc("/self/get-info", h.MwWithAuth(h.MwLogging(h.GetUserSelfInfo))).Methods(http.MethodGet)

	h.router.HandleFunc("/users/get-by-chat", h.MwWithAuth(h.MwLogging(h.GetUsersByChat))).Methods(http.MethodGet)
	h.router.HandleFunc("/users/add-to-chat", h.MwWithAuth(h.MwLogging(h.AddUserToChat))).Methods(http.MethodPost)
	h.router.HandleFunc("/users/delete-from-chat", h.MwWithAuth(h.MwLogging(h.DeleteUserFromChat))).Methods(http.MethodPost)

	h.router.HandleFunc("/chats/create", h.MwWithAuth(h.MwLogging(h.CreateChat))).Methods(http.MethodPost)
	h.router.HandleFunc("/chats/update", h.MwWithAuth(h.MwLogging(h.UpdateChat))).Methods(http.MethodPost)
	h.router.HandleFunc("/chats/delete", h.MwWithAuth(h.MwLogging(h.DeleteChat))).Methods(http.MethodPost)
	h.router.HandleFunc("/chats/get-my", h.MwWithAuth(h.MwLogging(h.GetUserChats))).Methods(http.MethodGet)

	h.router.HandleFunc("/messages/create", h.MwWithAuth(h.MwLogging(h.CreateMessage))).Methods(http.MethodPost)
	h.router.HandleFunc("/messages/update", h.MwWithAuth(h.MwLogging(h.UpdateMessage))).Methods(http.MethodPost)
	h.router.HandleFunc("/messages/delete", h.MwWithAuth(h.MwLogging(h.DeleteMessage))).Methods(http.MethodPost)
	h.router.HandleFunc("/messages/get-by-chat", h.MwWithAuth(h.MwLogging(h.GetMessages))).Methods(http.MethodGet)
	h.router.HandleFunc("/messages/min-id-in-chat", h.MwWithAuth(h.MwLogging(h.GetMessages))).Methods(http.MethodGet)
	h.router.HandleFunc("/messages/ws", h.MwWithAuth(h.MwLogging(h.eventHandler.HandleWS)))
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}
