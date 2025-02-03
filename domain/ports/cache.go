package ports

import (
	"context"
	"messanger/pkg/errors"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, v int, ttl time.Duration) *errors.Error
	Get(ctx context.Context, key string) (int, *errors.Error)
	Del(ctx context.Context, key string) *errors.Error
	TTL(ctx context.Context, key string) time.Duration
}
