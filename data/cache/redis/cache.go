package cache

import (
	"context"
	errorsutils "errors"
	"github.com/redis/go-redis/v9"
	"messanger/domain/models"
	"messanger/pkg/errors"
	"net/http"
	"strconv"
	"time"
)

type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{client}
}

func (c *Cache) Set(ctx context.Context, key string, v int, ttl time.Duration) *errors.Error {
	if err := c.client.Set(ctx, key, v, ttl).Err(); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (c *Cache) Get(ctx context.Context, key string) (int, *errors.Error) {
	res := c.client.Get(ctx, key)
	if err := res.Err(); err != nil {
		if errorsutils.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}

	v, _ := strconv.Atoi(res.Val())

	return v, nil
}

func (c *Cache) Del(ctx context.Context, key string) *errors.Error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return errors.New(err, models.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (c *Cache) TTL(ctx context.Context, key string) time.Duration {
	return c.client.TTL(ctx, key).Val()
}
