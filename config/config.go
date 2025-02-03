package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"reflect"
)

type Config struct {
	HttpServer  *HttpServerConfig  `json:"http_server" yaml:"http_server"`
	AuthService *AuthServiceConfig `json:"auth_service" yaml:"auth_service"`
	Redis       *RedisConfig       `json:"redis" yaml:"redis"`
	MySQL       *MySQLConfig       `json:"mysql" yaml:"mysql"`
}

type HttpServerConfig struct {
	Addr            string `json:"addr" yaml:"addr"`
	ReadTimeoutSec  int    `json:"read_timeout_sec" yaml:"read_timeout_sec"`
	WriteTimeoutSec int    `json:"write_timeout_sec" yaml:"write_timeout_sec"`
}

type AuthServiceConfig struct {
	AccessTokenTTLMin    int    `json:"access_token_min" yaml:"access_token_ttl_min"`
	RefreshTokenTTLDays  int    `json:"refresh_token_ttl_days" yaml:"refresh_token_ttl_days"`
	AccessTokenSignKey   string `json:"access_token_sign_key" yaml:"access_token_sign_key"`
	DurationBlockUserMin int    `json:"duration_block_user_min" yaml:"duration_block_user_min"`
	LoginAttempts        int    `json:"login_attempts" yaml:"login_attempts"`
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
