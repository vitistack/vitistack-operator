package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// entryTTL is how long a cached value lives after it was last set
const entryTTL = 6 * time.Hour

var Cache *VitistackCache

// storage is the key/value layer behind VitistackCache
type storage interface {
	Get(key string) (any, bool)
	Set(key string, value any)
	Keys() []string
	Remove(key string)
}

type VitistackCache struct {
	cacheLayer storage
}

func (dccache VitistackCache) NewVitistackCache() (*VitistackCache, error) {
	dccache = VitistackCache{
		cacheLayer: newMemoryStore(entryTTL),
	}
	return &dccache, nil
}

func (dccache VitistackCache) Get(ctx context.Context, key string) (string, error) {
	value, _ := dccache.cacheLayer.Get(key)
	if value == nil {
		return "", nil
	}
	s, ok := value.(string)
	if !ok {
		return "", errors.New("cached value is not a string")
	}
	return s, nil
}

func (dccache VitistackCache) Set(ctx context.Context, key string, value any) error {
	stringvalue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	dccache.cacheLayer.Set(key, string(stringvalue))
	return nil
}

// Delete removes key from the cache. Deleting a missing key is not an error.
func (dccache VitistackCache) Delete(ctx context.Context, key string) error {
	dccache.cacheLayer.Remove(key)
	return nil
}

func (dccache VitistackCache) Keys(ctx context.Context) ([]string, error) {
	keys := dccache.cacheLayer.Keys()
	if len(keys) == 0 {
		return nil, errors.New("no keys found")
	}
	return keys, nil
}

func (dccache VitistackCache) GetByKey(ctx context.Context, key string) (string, error) {
	value, ok := dccache.cacheLayer.Get(key)
	if !ok {
		return "", errors.New("key not found")
	}
	if value == nil {
		return "", nil
	}
	s, ok := value.(string)
	if !ok {
		return "", errors.New("cached value is not a string")
	}
	return s, nil
}
