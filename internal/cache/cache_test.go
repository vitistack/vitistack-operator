package cache

import (
	"context"
	"testing"
)

func newTestVitistackCache(t *testing.T) *VitistackCache {
	t.Helper()
	c, err := VitistackCache{}.NewVitistackCache()
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	return c
}

func TestVitistackCache_SetThenGet_ReturnsJSON(t *testing.T) {
	c := newTestVitistackCache(t)
	ctx := context.Background()

	if err := c.Set(ctx, "key", map[string]any{"name": "provider-a"}); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := c.Get(ctx, "key")
	if err != nil || got != `{"name":"provider-a"}` {
		t.Fatalf(`expected ({"name":"provider-a"}, nil), got (%q, %v)`, got, err)
	}
}

// DeleteResource only publishes the DELETE event when Cache.Delete succeeds, so deleting
// a key that was never cached, or has already expired, must not fail.
func TestVitistackCache_DeleteMissingKey_Succeeds(t *testing.T) {
	c := newTestVitistackCache(t)

	if err := c.Delete(context.Background(), "missing"); err != nil {
		t.Fatalf("expected deleting a missing key to succeed, got %v", err)
	}
}
