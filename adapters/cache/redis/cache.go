package cache

import (
	"context"
	errorsutils "errors"
	"github.com/redis/go-redis/v9"
	"messanger/config"
	"messanger/domain"
	"messanger/pkg/errors"
	"net/http"
	"strconv"
	"time"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCache(cfg *config.RedisConfig) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &Cache{
		client: client,
		ttl:    time.Minute * 5,
	}, nil
}

func (c *Cache) SetTTL(ttl time.Duration) {
	c.ttl = ttl
}

func (c *Cache) Set(key string, v int) *errors.Error {
	if err := c.client.Set(context.Background(), key, v, c.ttl).Err(); err != nil {
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
