package dynamichandler

import (
	"testing"

	localcache "github.com/vitistack/vitistack-operator/internal/cache"
	"github.com/vitistack/vitistack-operator/pkg/eventmanager"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func newTestObject(kind, resourceVersion string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "vitistack.io/v1alpha1",
		"kind":       kind,
		"metadata": map[string]any{
			"name":            "object-a",
			"uid":             "3f1c2d4e-0000-4000-8000-000000000001",
			"resourceVersion": resourceVersion,
		},
	}}
}

// Subscribers need the previous object to tell a meaningful change from a status
// heartbeat without asking the API server.
func TestUpdateResource_PublishesPreviousObject(t *testing.T) {
	previousCache := localcache.Cache
	c, err := localcache.VitistackCache{}.NewVitistackCache()
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	localcache.Cache = c
	t.Cleanup(func() { localcache.Cache = previousCache })

	const kind = "UpdateResourceTestKind"
	var published eventmanager.ResourceEvent
	eventmanager.EventBus.Subscribe(kind, func(event eventmanager.ResourceEvent) { published = event })

	old := newTestObject(kind, "1")
	updated := newTestObject(kind, "2")
	NewDynamicClientHandler().UpdateResource(old, updated)

	if published.Resource != updated {
		t.Fatalf("expected Resource to be the updated object, got %v", published.Resource)
	}
	if published.OldResource != old {
		t.Fatalf("expected OldResource to be the previous object, got %v", published.OldResource)
	}
}
