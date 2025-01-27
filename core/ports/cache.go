package ports

import (
	"messanger/pkg/errors"
	"time"
)

type Cache interface {
	Set(key string, v int, ttl time.Duration) *errors.Error
	Get(key string) (int, *errors.Error)
	Del(key string) *errors.Error
}
