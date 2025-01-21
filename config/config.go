package config

import "time"

type Config struct {
}

type AuthServiceConfig struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type RedisConfig struct {
	Host     string `json:"host"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}
