package ports

import (
	"messanger/pkg/errors"
	"time"
)

type TokenCache interface {
	SetTTL(ttl time.Duration)
	Set(key string, v int) *errors.Error
	Get(key string) (int, *errors.Error)
	Del(key string) *errors.Error
}
