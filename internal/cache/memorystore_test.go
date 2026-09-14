package cache

import (
	"testing"
	"time"
)

const testTTL = time.Hour

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time          { return c.now }
func (c *fakeClock) Advance(d time.Duration) { c.now = c.now.Add(d) }

func newTestMemoryStore() (*memoryStore, *fakeClock) {
	clock := &fakeClock{now: time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)}
	store := newMemoryStore(testTTL)
	store.now = clock.Now
	return store, clock
}

func TestMemoryStore_SetThenGet(t *testing.T) {
	store, _ := newTestMemoryStore()

	store.Set("key", "value")

	if value, ok := store.Get("key"); !ok || value != "value" {
		t.Fatalf("expected (value, true), got (%v, %v)", value, ok)
	}
}

func TestMemoryStore_SetOverwrites(t *testing.T) {
	store, _ := newTestMemoryStore()

	store.Set("key", "first")
	store.Set("key", "second")

	if value, _ := store.Get("key"); value != "second" {
		t.Fatalf("expected the latest value, got %v", value)
	}
}

func TestMemoryStore_EntryExpiresAfterTTL(t *testing.T) {
	store, clock := newTestMemoryStore()
	store.Set("key", "value")

	clock.Advance(testTTL + time.Second)

	if value, ok := store.Get("key"); ok {
		t.Fatalf("expected expired entry to be a miss, got %v", value)
	}
	if keys := store.Keys(); len(keys) != 0 {
		t.Fatalf("expected no keys after expiry, got %v", keys)
	}
}

// Informer resyncs re-set every live object, which must keep it from expiring.
func TestMemoryStore_SetRenewsTTL(t *testing.T) {
	store, clock := newTestMemoryStore()
	store.Set("key", "value")

	clock.Advance(testTTL - time.Minute)
	store.Set("key", "value")
	clock.Advance(2 * time.Minute)

	if _, ok := store.Get("key"); !ok {
		t.Fatal("expected re-set entry to still be cached")
	}
}

// Expired entries must be freed, not just hidden, or deleted objects' entries pile up.
func TestMemoryStore_LaterWriteRemovesExpiredEntries(t *testing.T) {
	store, clock := newTestMemoryStore()
	store.Set("expired", "value")

	clock.Advance(testTTL + sweepInterval)
	store.Set("fresh", "value")

	if _, found := store.entries["expired"]; found {
		t.Fatalf("expected expired entry to be removed from memory, entries: %v", store.entries)
	}
}
