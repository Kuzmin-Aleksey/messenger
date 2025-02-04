package http

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"messanger/domain/service/auth"
	"messanger/domain/service/chats"
	"messanger/domain/service/groups"
	messages2 "messanger/domain/service/messages"
	"messanger/domain/service/users"
	"messanger/pkg/errors"
	"net/http"
)

type Logger interface {
	Println(v ...any)
	Printf(format string, v ...any)
}

type Handler struct {
	router *mux.Router

	auth     *auth.AuthService
	users    *users.UsersService
	messages *messages2.MessagesService
	chats    *chats.ChatService
	groups   *groups.GroupService

	logger Logger
	info   *HttpLogger

	connManager *messages2.ConnectionsManager
	wsUpgrader  *websocket.Upgrader
}

func NewHandler(
	auth *auth.AuthService,
	users *users.UsersService,
	messages *messages2.MessagesService,
	chats *chats.ChatService,
	groups *groups.GroupService,

	logger Logger,
	// info io.Writer,
) *Handler {
	h := &Handler{
		auth:     auth,
		users:    users,
		messages: messages,
		chats:    chats,
		groups:   groups,

		logger:      logger,
		info:        NewHttpLogger(),
		router:      mux.NewRouter(),
		connManager: messages.NewConnectionsManager(),
		wsUpgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
	h.wsUpgrader.Error = func(w http.ResponseWriter, _ *http.Request, status int, reason error) {
		h.writeJSONError(w, errors.New(reason, reason, status))
	}
	return h
}

func (h *Handler) InitRouter() {
	h.router.HandleFunc("/register", h.MwLogging(h.Register)).Methods(http.MethodPost)
	h.router.HandleFunc("/auth/login/1fa", h.MwLogging(h.Login1FA)).Methods(http.MethodPost)
	h.router.HandleFunc("/auth/login/2fa", h.MwLogging(h.Login2FA)).Methods(http.MethodPost)
	h.router.HandleFunc("/auth/refresh-tokens", h.MwLogging(h.UpdateTokens)).Methods(http.MethodPost)

	h.router.HandleFunc("/self/update", h.MwLogging(h.MwWithAuth(h.UpdateUser))).Methods(http.MethodPost)
	h.router.HandleFunc("/self/delete", h.MwLogging(h.MwWithAuth(h.DeleteUser))).Methods(http.MethodPost)
	h.router.HandleFunc("/self/set-show-phone", h.MwLogging(h.MwWithAuth(h.SetShowPhone))).Methods(http.MethodPost)

	h.router.HandleFunc("/users/check-username", h.MwLogging(h.CheckUsername)).Methods(http.MethodGet)
	h.router.HandleFunc("/users/status", h.MwLogging(h.MwWithAuth(h.CheckOnline))).Methods(http.MethodGet)
	h.router.HandleFunc("/users/find", h.MwLogging(h.MwWithAuth(h.FindUser))).Methods(http.MethodGet)
	h.router.HandleFunc("/users/create-chat", h.MwLogging(h.MwWithAuth(h.CreateChatWithUser))).Methods(http.MethodPost)

	h.router.HandleFunc("/chats/get-my", h.MwLogging(h.MwWithAuth(h.GetAllUserChats))).Methods(http.MethodGet)

	h.router.HandleFunc("/groups/create", h.MwLogging(h.MwWithAuth(h.CreateGroup))).Methods(http.MethodPost)
	h.router.HandleFunc("/groups/update", h.MwLogging(h.MwWithAuth(h.UpdateGroup))).Methods(http.MethodPost)
	h.router.HandleFunc("/groups/delete", h.MwLogging(h.MwWithAuth(h.DeleteGroup))).Methods(http.MethodPost)
	h.router.HandleFunc("/groups/get-users", h.MwLogging(h.MwWithAuth(h.GetUsersByGroup))).Methods(http.MethodGet)
	h.router.HandleFunc("/groups/add-user", h.MwLogging(h.MwWithAuth(h.AddUserToGroup))).Methods(http.MethodPost)
	h.router.HandleFunc("/groups/delete-user", h.MwLogging(h.MwWithAuth(h.DeleteUserFromGroup))).Methods(http.MethodPost)
	h.router.HandleFunc("/groups/set-role", h.MwLogging(h.MwWithAuth(h.SetUsersRoleInGroup))).Methods(http.MethodPost)

	h.router.HandleFunc("/contacts/add", h.MwLogging(h.MwWithAuth(h.AddContact))).Methods(http.MethodPost)
	h.router.HandleFunc("/contacts/rename", h.MwLogging(h.MwWithAuth(h.RenameContact))).Methods(http.MethodPost)
	h.router.HandleFunc("/contacts/get-all", h.MwLogging(h.MwWithAuth(h.GetUserContacts))).Methods(http.MethodGet)
	h.router.HandleFunc("/contacts/delete", h.MwLogging(h.MwWithAuth(h.DeleteContact))).Methods(http.MethodPost)

	h.router.HandleFunc("/messages/create", h.MwLogging(h.MwWithAuth(h.CreateMessage))).Methods(http.MethodPost)
	h.router.HandleFunc("/messages/update", h.MwLogging(h.MwWithAuth(h.UpdateMessage))).Methods(http.MethodPost)
	h.router.HandleFunc("/messages/delete", h.MwLogging(h.MwWithAuth(h.DeleteMessage))).Methods(http.MethodPost)
	h.router.HandleFunc("/messages/get-by-chat", h.MwLogging(h.MwWithAuth(h.GetMessages))).Methods(http.MethodGet)
	h.router.HandleFunc("/messages/min-id-in-chat", h.MwLogging(h.MwWithAuth(h.GetMinMassageIdInChat))).Methods(http.MethodGet)
	h.router.HandleFunc("/messages/ws", h.MwLogging(h.MwWithAuth(h.HandleWS)))
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}
