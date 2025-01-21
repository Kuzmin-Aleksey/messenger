package ports

import "time"

type TokenCache interface {
	SetTTL(ttl time.Duration)
	Set(key string, v int) error
	Get(key string) (int, error)
	Del(key string) error
}
