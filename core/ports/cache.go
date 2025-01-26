package ports

import (
	"messanger/pkg/errors"
	"time"
)

type Cache interface {
	Set(key any, v int, ttl time.Duration) *errors.Error
	Get(key any) (int, *errors.Error)
	Del(key any) *errors.Error
}
