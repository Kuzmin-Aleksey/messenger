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

	auth         *service.AuthService
	users        *service.UsersService
	messages     *service.MessagesService
	chats        *service.ChatService
	emailService *service.EmailService

	errors Logger
	info   *HttpLogger

	eventHandler *EventHandler
}

func NewHandler(
	auth *service.AuthService,
	users *service.UsersService,
	messages *service.MessagesService,
	chats *service.ChatService,
	emailService *service.EmailService,

	errors Logger,
	// info io.Writer,
) *Handler {
	return &Handler{
		auth:         auth,
		users:        users,
		messages:     messages,
		chats:        chats,
		errors:       errors,
		emailService: emailService,
		info:         NewHttpLogger(),
		router:       mux.NewRouter(),
		eventHandler: NewEventHandler(errors),
	}
}

func (h *Handler) InitRouter() {
	h.router.HandleFunc("/auth/register", h.MwLogging(h.Register)).Methods(http.MethodPost)
	h.router.HandleFunc("/auth/login", h.MwLogging(h.Login)).Methods(http.MethodPost)
	h.router.HandleFunc("/auth/refresh-tokens", h.MwLogging(h.UpdateTokens)).Methods(http.MethodPost)

	h.router.HandleFunc("/self/update", h.MwLogging(h.MwWithAuth(h.UpdateUserSelf))).Methods(http.MethodPost)
	h.router.HandleFunc("/self/delete", h.MwLogging(h.MwWithAuth(h.DeleteUserSelf))).Methods(http.MethodPost)
	h.router.HandleFunc("/self/info", h.MwLogging(h.MwWithAuth(h.GetUserSelfInfo))).Methods(http.MethodGet)

	h.router.HandleFunc("/confirm-email", h.MwLogging(h.ConfirmEmail))

	h.router.HandleFunc("/users/get-by-chat", h.MwLogging(h.MwWithAuth(h.GetUsersByChat))).Methods(http.MethodGet)
	h.router.HandleFunc("/users/add-to-chat", h.MwLogging(h.MwWithAuth(h.AddUserToChat))).Methods(http.MethodPost)
	h.router.HandleFunc("/users/set-role", h.MwLogging(h.MwWithAuth(h.SetRole))).Methods(http.MethodPost)
	h.router.HandleFunc("/users/delete-from-chat", h.MwLogging(h.MwWithAuth(h.DeleteUserFromChat))).Methods(http.MethodPost)

	h.router.HandleFunc("/chats/create", h.MwLogging(h.MwWithAuth(h.CreateChat))).Methods(http.MethodPost)
	h.router.HandleFunc("/chats/update", h.MwLogging(h.MwWithAuth(h.UpdateChat))).Methods(http.MethodPost)
	h.router.HandleFunc("/chats/delete", h.MwLogging(h.MwWithAuth(h.DeleteChat))).Methods(http.MethodPost)
	h.router.HandleFunc("/chats/get-my", h.MwLogging(h.MwWithAuth(h.GetUserChats))).Methods(http.MethodGet)

	h.router.HandleFunc("/messages/create", h.MwLogging(h.MwWithAuth(h.CreateMessage))).Methods(http.MethodPost)
	h.router.HandleFunc("/messages/update", h.MwLogging(h.MwWithAuth(h.UpdateMessage))).Methods(http.MethodPost)
	h.router.HandleFunc("/messages/delete", h.MwLogging(h.MwWithAuth(h.DeleteMessage))).Methods(http.MethodPost)
	h.router.HandleFunc("/messages/get-by-chat", h.MwLogging(h.MwWithAuth(h.GetMessages))).Methods(http.MethodGet)
	h.router.HandleFunc("/messages/min-id-in-chat", h.MwLogging(h.MwWithAuth(h.GetMinMassageIdInChat))).Methods(http.MethodGet)
	h.router.HandleFunc("/messages/ws", h.MwLogging(h.MwWithAuth(h.MwWithHijacker(h.eventHandler.HandleWS))))
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}
