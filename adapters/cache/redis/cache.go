package cache

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"messanger/config"
	tr "messanger/pkg/error_trace"
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
		return nil, tr.Trace(err)
	}

	return &Cache{
		client: client,
		ttl:    time.Minute * 5,
	}, nil
}

func (c *Cache) SetTTL(ttl time.Duration) {
	c.ttl = ttl
}

func (c *Cache) Set(key string, v int) error {
	if err := c.client.Set(context.Background(), key, v, c.ttl).Err(); err != nil {
		return tr.Trace(err)
	}
	return nil
}

func (c *Cache) Get(key string) (int, error) {
	res := c.client.Get(context.Background(), key)
	if err := res.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}

	v, _ := strconv.Atoi(res.Val())

	return v, nil
}

func (c *Cache) Del(key string) error {
	if err := c.client.Del(context.Background(), key).Err(); err != nil {
		return tr.Trace(err)
	}
	return nil
}
