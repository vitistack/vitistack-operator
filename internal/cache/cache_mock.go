package cache

import (
	"context"
	"errors"
	"sync"
)

// Mock implementation for testing
func NewMockVitistackCache() *VitistackCache {
	mockCache := &VitistackCache{}
	mockCache.cacheLayer = &mockCacheLayer{
		data: make(map[string]any),
	}
	return mockCache
}

type mockCacheLayer struct {
	data map[string]any
	mu   sync.RWMutex
}

func (m *mockCacheLayer) Get(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.data[key]
	return value, exists
}

func (m *mockCacheLayer) Set(key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *mockCacheLayer) Keys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]string, 0, len(m.data))
	for key := range m.data {
		keys = append(keys, key)
	}
	return keys
}

func (m *mockCacheLayer) Remove(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

type MockCache struct {
	data map[string]any
	mu   sync.RWMutex
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]any),
	}
}

func (m *MockCache) Get(ctx context.Context, key string) (any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if value, exists := m.data[key]; exists {
		return value, nil
	}
	return nil, errors.New("key not found")
}

func (m *MockCache) Set(ctx context.Context, key string, value any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[key]; exists {
		delete(m.data, key)
		return nil
	}
	return errors.New("key not found")
}
