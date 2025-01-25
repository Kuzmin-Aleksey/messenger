package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"reflect"
)

type Config struct {
	HttpServer  *HttpServerConfig   `yaml:"http_server"`
	AuthService *AuthServiceConfig  `yaml:"auth_service"`
	Email       *EmailServiceConfig `yaml:"email_service"`
	Redis       *RedisConfig        `yaml:"redis"`
	MySQL       *MySQLConfig        `yaml:"mysql"`
}

type HttpServerConfig struct {
	Addr            string `json:"addr" yaml:"addr"`
	ReadTimeoutSec  int    `json:"read_timeout_sec" yaml:"read_timeout_sec"`
	WriteTimeoutSec int    `json:"write_timeout_sec" yaml:"write_timeout_sec"`
}

type AuthServiceConfig struct {
	AccessTokenTTLMin   int    `json:"access_token_min" yaml:"access_token_ttl_min"`
	RefreshTokenTTLDays int    `json:"refresh_token_ttl_days" yaml:"refresh_token_ttl_days"`
	AccessTokenSignKey  string `json:"access_token_sign_key" yaml:"access_token_sign_key"`
}

type RedisConfig struct {
	Host     string `json:"host" yaml:"host"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"`
}

type MySQLConfig struct {
	Host              string `json:"host" yaml:"host"`
	Username          string `json:"username" yaml:"username"`
	Password          string `json:"password" yaml:"password"`
	Schema            string `json:"schema" yaml:"schema"`
	ConnectTimeoutSec int    `json:"connect_timeout_sec" yaml:"connect_timeout_sec"`
}

type EmailServiceConfig struct {
	TTLMin          int             `json:"ttl_min" yaml:"ttl_min"`
	TokenSignKey    string          `json:"token_sign_key" yaml:"token_sign_key"`
	ConfirmEmailURL string          `json:"confirm_email_url" yaml:"confirm_email_url"`
	SMTP            EmailSMTPConfig `json:"smtp" yaml:"smtp"`
}

type EmailSMTPConfig struct {
	Server   string `json:"server" yaml:"server"`
	Port     int    `json:"port" yaml:"port"`
	Email    string `json:"email" yaml:"email"`
	Password string `json:"password" yaml:"password"`
}

func GetConfig(path string) (*Config, error) {
	cfg := new(Config)
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, err
	}

	v := reflect.Indirect(reflect.ValueOf(*cfg))
	fields := reflect.VisibleFields(reflect.TypeOf(*cfg))
	for _, field := range fields {
		f := v.FieldByName(field.Name)
		if f.IsNil() {
			return nil, fmt.Errorf("config %s not found", field.Tag.Get("yaml"))
		}
	}

	return cfg, nil
}
