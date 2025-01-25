package http_server

import (
	"messanger/config"
	"net/http"
	"time"
)

func NewHttpServer(h http.Handler, cfg *config.HttpServerConfig) *http.Server {
	return &http.Server{
		Handler:      h,
		Addr:         cfg.Addr,
		ReadTimeout:  time.Duration(cfg.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeoutSec) * time.Second,
	}
}
