package http

import (
	"net/http"
)

func (h *Handler) GetAllUserChats(w http.ResponseWriter, r *http.Request) {
	chats, err := h.chats.GetAllUserChats(r.Context())
	if err != nil {
		h.writeJSONError(w, err.Trace())
		return
	}
	h.writeJSON(w, http.StatusOK, chats)
}
