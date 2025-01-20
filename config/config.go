package config

import "time"

type Config struct {
}

type AuthServiceConfig struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}
