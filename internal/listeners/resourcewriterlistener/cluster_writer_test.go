package resourcewriterlistener

import (
	"testing"

	"github.com/vitistack/vitistack-operator/pkg/eventmanager"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

// newTestKubernetesCluster builds a KubernetesCluster the way the informer delivers it.
// lastUpdated mimics the talos-operator heartbeat, which changes on every reconcile
// without touching anything the vitistack status tracks.
func newTestKubernetesCluster(resourceVersion, phase, lastUpdated string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "vitistack.io/v1alpha1",
		"kind":       "KubernetesCluster",
		"metadata": map[string]any{
			"name":            "cluster-a",
			"namespace":       "tenant-a",
			"resourceVersion": resourceVersion,
		},
		"spec": map[string]any{
			"topology": map[string]any{
				"controlplane": map[string]any{"replicas": int64(3)},
				"workers":      []any{map[string]any{"replicas": int64(2)}},
			},
		},
		"status": map[string]any{
			"phase": phase,
			"state": map[string]any{"lastUpdated": lastUpdated},
		},
	}}
}

// Every KubernetesCluster status heartbeat used to cost two GETs of the Vitistack CR.
// At ~40 heartbeats/s that outran the handler and the informer queue grew until OOM.
func TestKubernetesClusterUpdate_HeartbeatOnly_MakesNoAPICalls(t *testing.T) {
	client := useFakeDynamicClient(t, newTestVitistack())
	old := newTestKubernetesCluster("100", "Ready", "2026-09-14T12:00:00Z")
	updated := newTestKubernetesCluster("101", "Ready", "2026-09-14T12:00:03Z")

	handleKubernetesClusterEvents(eventmanager.ResourceEvent{
		Type:        eventmanager.EventUpdate,
		Resource:    updated,
		OldResource: old,
	})

	if actions := client.Actions(); len(actions) != 0 {
		t.Fatalf("heartbeat-only update should not call the API, got %d calls: %v", len(actions), actions)
	}
}

func TestKubernetesClusterUpdate_PhaseChanged_UpdatesVitistackStatus(t *testing.T) {
	client := useFakeDynamicClient(t, newTestVitistack())
	old := newTestKubernetesCluster("100", "Provisioning", "2026-09-14T12:00:00Z")
	updated := newTestKubernetesCluster("101", "Ready", "2026-09-14T12:00:03Z")

	handleKubernetesClusterEvents(eventmanager.ResourceEvent{
		Type:        eventmanager.EventUpdate,
		Resource:    updated,
		OldResource: old,
	})

	assertVitistackClusterPhase(t, client, "Ready")
}

// Informer resyncs redeliver the cached object unchanged (same resourceVersion).
// They must still reconcile, so a status write that failed earlier gets retried.
func TestKubernetesClusterResync_ReconcilesVitistackStatus(t *testing.T) {
	client := useFakeDynamicClient(t, newTestVitistack())
	cluster := newTestKubernetesCluster("100", "Ready", "2026-09-14T12:00:00Z")

	handleKubernetesClusterEvents(eventmanager.ResourceEvent{
		Type:        eventmanager.EventUpdate,
		Resource:    cluster,
		OldResource: cluster,
	})

	assertVitistackClusterPhase(t, client, "Ready")
}

func assertVitistackClusterPhase(t *testing.T, client *dynamicfake.FakeDynamicClient, want string) {
	t.Helper()
	clusters, _ := getVitistackStatus(t, client)["clusters"].([]any)
	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster in vitistack status, got %d: %v", len(clusters), clusters)
	}
	cluster, _ := clusters[0].(map[string]any)
	if cluster["phase"] != want {
		t.Fatalf("expected cluster phase %q in vitistack status, got %v", want, cluster["phase"])
	}
}
