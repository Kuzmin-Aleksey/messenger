package http

import (
	"fmt"
	"github.com/gorilla/websocket"
	"messanger/domain/service/auth"
	"messanger/domain/service/messages"
	"messanger/pkg/errors"
	"net/http"
)

type WsConnAdapter struct {
	conn   *websocket.Conn
	logger Logger
}

func (c *WsConnAdapter) Ping() bool {
	if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
		c.logger.Println(errors.Trace(fmt.Errorf("ws: ping error: %w", err)))
		return false
	}
	return true
}

func (c *WsConnAdapter) Send(event *messages.Event) bool {
	if err := c.conn.WriteJSON(event); err != nil {
		c.logger.Println(errors.Trace(fmt.Errorf("ws: write json error: %w", err)))
		return false
	}
	return true
}

type WsHandler struct {
	connManager *messages.ConnectionsManager
	wsUpgrader  *websocket.Upgrader
	logger      Logger
}

func (h *Handler) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := h.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			return // error already written
		}
		h.writeJSONError(w, errors.New(err, "upgrade to ws failed", http.StatusBadRequest))
		return
	}
	userId := auth.ExtractUser(r.Context())

	if err := h.connManager.InsertConn(r.Context(), userId, &WsConnAdapter{
		conn:   conn,
		logger: h.logger,
	}); err != nil {
		h.writeJSONError(w, err)
	}
}
