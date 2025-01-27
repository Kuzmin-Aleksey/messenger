package cache

import (
	"context"
	errorsutils "errors"
	"github.com/redis/go-redis/v9"
	"messanger/domain"
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

func (c *Cache) Set(key string, v int, ttl time.Duration) *errors.Error {
	if err := c.client.Set(context.Background(), key, v, ttl).Err(); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}

func (c *Cache) Get(key string) (int, *errors.Error) {
	res := c.client.Get(context.Background(), key)
	if err := res.Err(); err != nil {
		if errorsutils.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}

	v, _ := strconv.Atoi(res.Val())

	return v, nil
}

func (c *Cache) Del(key string) *errors.Error {
	if err := c.client.Del(context.Background(), key).Err(); err != nil {
		return errors.New(err, domain.ErrDatabaseError, http.StatusInternalServerError)
	}
	return nil
}
