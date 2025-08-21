package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/NorskHelsenett/ror/pkg/helpers/kvcachehelper"
	"github.com/NorskHelsenett/ror/pkg/helpers/kvcachehelper/memorycache"
)

var Cache *VitistackCache

type VitistackCache struct {
	cacheLayer kvcachehelper.CacheInterface
}

func (dccache VitistackCache) NewVitistackCache() (*VitistackCache, error) {
	dccache = VitistackCache{
		cacheLayer: memorycache.NewKvCache(kvcachehelper.CacheOptions{
			Timeout: time.Hour * 6,
		}),
	}
	return &dccache, nil
}

func (dccache VitistackCache) Get(ctx context.Context, key string) (string, error) {
	value, _ := dccache.cacheLayer.Get(ctx, key)
	return value, nil
}

func (dccache VitistackCache) Set(ctx context.Context, key string, value any) error {
	stringvalue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	dccache.cacheLayer.Set(ctx, key, string(stringvalue))
	return nil
}

func (dccache VitistackCache) Delete(ctx context.Context, key string) error {
	ok := dccache.cacheLayer.Remove(ctx, key)
	if !ok {
		return errors.New("could not delete key")
	}
	return nil
}

func (dccache VitistackCache) Keys(ctx context.Context) ([]string, error) {
	keys, err := dccache.cacheLayer.Keys(ctx)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, errors.New("no keys found")
	}
	return keys, nil
}

func (dccache VitistackCache) GetByKey(ctx context.Context, key string) (string, error) {
	value, ok := dccache.cacheLayer.Get(ctx, key)
	if !ok {
		return "", errors.New("key not found")
	}
	return value, nil
}
