package cache

import (
	"sync"
	"time"
)

// sweepInterval is the minimum time between passes that free expired entries.
const sweepInterval = time.Minute

// memoryStore is an in-memory key/value store. An entry expires ttl after it was last set.
// Expired entries are hidden right away and freed by the first write after sweepInterval,
// so the store needs no background goroutine.
type memoryStore struct {
	mu        sync.RWMutex
	entries   map[string]memoryEntry
	ttl       time.Duration
	now       func() time.Time
	lastSweep time.Time
}

type memoryEntry struct {
	value     any
	expiresAt time.Time
}

func newMemoryStore(ttl time.Duration) *memoryStore {
	return &memoryStore{
		entries: make(map[string]memoryEntry),
		ttl:     ttl,
		now:     time.Now,
	}
}

// Get returns the value stored under key, or false if it is missing or expired.
func (s *memoryStore) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[key]
	if !ok || s.now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.value, true
}

// Set stores value under key, replacing any previous value and restarting its ttl.
func (s *memoryStore) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if now.Sub(s.lastSweep) >= sweepInterval {
		s.removeExpired(now)
		s.lastSweep = now
	}
	s.entries[key] = memoryEntry{value: value, expiresAt: now.Add(s.ttl)}
}

// Keys returns the keys of all entries that have not expired.
func (s *memoryStore) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.now()
	keys := make([]string, 0, len(s.entries))
	for key, entry := range s.entries {
		if !now.After(entry.expiresAt) {
			keys = append(keys, key)
		}
	}
	return keys
}

// Remove deletes the entry for key. Removing a missing key is not an error.
func (s *memoryStore) Remove(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
}

// removeExpired frees expired entries. The caller must hold the write lock.
func (s *memoryStore) removeExpired(now time.Time) {
	for key, entry := range s.entries {
		if now.After(entry.expiresAt) {
			delete(s.entries, key)
		}
	}
}
