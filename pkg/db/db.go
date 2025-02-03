package db

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"messanger/config"
	"messanger/pkg/errors"
	"time"
)

func Connect(cfg *config.MySQLConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", cfg.Username, cfg.Password, cfg.Host, cfg.Schema))
	if err != nil {
		return nil, errors.Trace(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnectTimeoutSec)*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, errors.Trace(err)
	}
	return db, nil
}
