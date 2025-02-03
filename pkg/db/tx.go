package db

import (
	"context"
	"database/sql"
	"fmt"
	"messanger/domain/models"
	"messanger/pkg/errors"
	"net/http"
)

type Beginner interface {
	Begin() (*sql.Tx, error)
}

type txKey struct{}

func WithTx(ctx context.Context, db any) (context.Context, *errors.Error) {
	b, ok := db.(Beginner)
	if ok {
		tx, err := b.Begin()
		if err != nil {
			return nil, errors.New(fmt.Errorf("begin error: %w", err), models.ErrDatabaseError, http.StatusInternalServerError)
		}
		return context.WithValue(ctx, txKey{}, tx), nil
	}
	return ctx, nil
}

func Commit(ctx context.Context) *errors.Error {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if ok {
		if err := tx.Commit(); err != nil {
			return errors.New(fmt.Errorf("commit error: %w", err), models.ErrDatabaseError, http.StatusInternalServerError)
		}
	}
	return nil
}

func Rollback(ctx context.Context) *errors.Error {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if ok {
		if err := tx.Rollback(); err != nil {
			return errors.New(fmt.Errorf("rollback error: %w", err), models.ErrDatabaseError, http.StatusInternalServerError)
		}
	}
	return nil
}

func CommitOnDefer(ctx context.Context, err **errors.Error) {
	if *err != nil {
		if e := Rollback(ctx); e != nil {
			(*err).Msg = fmt.Sprintf("rollback error: %s; %s", e.Msg, (*err).Msg)
		}
		return
	}
	if e := Commit(ctx); e != nil {
		*err = errors.New(fmt.Errorf("commit error: %w", e), models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return
}

type DBWithTx struct {
	*sql.DB
}

func NewDBWithTx(db *sql.DB) *DBWithTx {
	return &DBWithTx{db}
}

func (db *DBWithTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if ok {
		return tx.QueryContext(ctx, query, args...)
	}
	return db.DB.QueryContext(ctx, query, args...)
}

func (db *DBWithTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if ok {
		return tx.QueryRowContext(ctx, query, args...)
	}
	return db.DB.QueryRowContext(ctx, query, args...)
}

func (db *DBWithTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if ok {
		return tx.ExecContext(ctx, query, args...)
	}
	return db.DB.ExecContext(ctx, query, args...)
}
