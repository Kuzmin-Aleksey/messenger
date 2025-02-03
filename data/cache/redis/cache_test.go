package cache

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSetGetDel(t *testing.T) {
	c, err := NewCache(cfg)
	require.NoError(t, err, "connect to redis failed")

	const testKey = "key"
	const testVal = 123

	err = c.Set(testKey, testVal)
	require.NoError(t, err)

	val, err := c.Get(testKey)
	require.NoError(t, err)
	require.Equal(t, testVal, val)

	err = c.Del(testKey)
	require.NoError(t, err)
	val, err = c.Get(testKey)
	require.NoError(t, err)
	require.Equal(t, 0, val)
}
