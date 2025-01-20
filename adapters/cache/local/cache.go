package cache

import "time"

type Cache struct {
	c   map[string]value
	ttl time.Duration
}

type value struct {
	v int
	t time.Time
}

func NewCache() *Cache {
	c := &Cache{
		c: make(map[string]value),
	}

	go func() {
		for {
			time.Sleep(time.Hour)
			now := time.Now()
			for k, v := range c.c {
				if now.After(v.t) {
					delete(c.c, k)
				}
			}
		}
	}()

	return c
}

func (c *Cache) SetTTL(ttl time.Duration) {
	c.ttl = ttl
}

func (c *Cache) Set(key string, v int) error {
	c.c[key] = value{
		v: v,
		t: time.Now().Add(c.ttl),
	}
	return nil
}

func (c *Cache) Get(key string) (int, bool, error) {
	v, ok := c.c[key]
	if ok {
		if time.Now().Before(v.t) {
			return v.v, ok, nil
		} else {
			delete(c.c, key)
		}
	}

	return 0, false, nil
}

func (c *Cache) Del(key string) error {
	delete(c.c, key)
	return nil
}
