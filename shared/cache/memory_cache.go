package cache

import (
	"context"
	"errors"
	"regexp"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type memoryCache struct {
	client *gocache.Cache
}

func NewMemoryCache() Cache {
	gocacheClient := gocache.New(5*time.Minute, 6*time.Minute)
	return &memoryCache{
		client: gocacheClient,
	}
}

func (m *memoryCache) Set(ctx context.Context, key string, value []byte, duration time.Duration) error {
	m.client.Set(key, value, duration)
	return nil
}

func (m *memoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	value, found := m.client.Get(key)
	if !found {
		return nil, errors.New("key not found")
	}
	return value, nil
}

func (m *memoryCache) Delete(ctx context.Context, key string) error {
	m.client.Delete(key)
	return nil
}

func (m *memoryCache) Clear() error {
	m.client.Flush()
	return nil
}

// DeleteByPattern deletes keys by regex pattern
func (m *memoryCache) DeleteByPattern(ctx context.Context, pattern string) error {
	for key := range m.client.Items() {
		if ok, _ := regexp.MatchString(pattern, key); ok {
			m.client.Delete(key)
		}
	}
	return nil
}
