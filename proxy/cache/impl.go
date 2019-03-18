package cache

import (
	"github.com/bradfitz/gomemcache/memcache"
	"time"
)

type cache struct {
	mclient *memcache.Client
}

func (c cache) Get(key string) (b []byte, err error) {
	item, err := c.mclient.Get(compressKey(key))
	if err != nil {
		return
	}

	b = item.Value
	return
}

func (c cache) Set(key string, value []byte) error {
	item := memcache.Item{
		Key:   compressKey(key),
		Value: value,
	}

	return c.mclient.Set(&item)
}

func compressKey(key string) string {
	// todo use md5 or whatever
	return key
}

func New(servers []string, timeout time.Duration) *cache {
	mclient := memcache.New(servers...)
	// little hack
	mclient.Timeout = timeout
	return &cache{mclient}
}
