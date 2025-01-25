package httpAPI

import (
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
)

func (h *Handler) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeJSONError(w, errors.New(err, domain.ErrParseForm, http.StatusBadRequest))
		return
	}
	token := r.Form.Get("token")

	if err := h.emailService.ConfirmUser(token); err != nil {
		h.writeJSONError(w, err)
	}
}
