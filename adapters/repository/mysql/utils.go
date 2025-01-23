package mysql

import (
	"database/sql"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
)

func Commit(tx *sql.Tx, e **errors.Error) {
	if *e != nil {
		tx.Rollback()
	} else {
		if err := tx.Commit(); err != nil {
			*e = errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
		}
	}
}
