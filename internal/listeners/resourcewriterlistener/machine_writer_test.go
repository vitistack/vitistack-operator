package resourcewriterlistener

import (
	"fmt"
	"testing"
	"time"

	"github.com/vitistack/vitistack-operator/pkg/eventmanager"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func newTestMachine(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "vitistack.io/v1alpha1",
		"kind":       "Machine",
		"metadata": map[string]any{
			"name":      name,
			"namespace": "tenant-a",
		},
	}}
}

// On startup the informer replays every existing Machine as an ADD. A full LIST per
// event is O(n²) work (707 LISTs of every machine on ptr-mgmt), so a burst of ADDs
// must collapse into a single recount.
func TestMachineAddBurst_ListsMachinesOnce(t *testing.T) {
	const machineCount = 50
	objs := []runtime.Object{newTestVitistack()}
	machines := make([]*unstructured.Unstructured, 0, machineCount)
	for i := range machineCount {
		machine := newTestMachine(fmt.Sprintf("machine-%d", i))
		machines = append(machines, machine)
		objs = append(objs, machine)
	}
	client := useFakeDynamicClient(t, objs...)

	previousDebounce := machineCountDebounce
	machineCountDebounce = 100 * time.Millisecond
	t.Cleanup(func() { machineCountDebounce = previousDebounce })

	for _, machine := range machines {
		handleMachineEvents(eventmanager.ResourceEvent{Type: eventmanager.EventAdd, Resource: machine})
	}

	eventually(t, 5*time.Second, func() bool {
		count, _, _ := unstructured.NestedInt64(getVitistackStatus(t, client), "activeMachines")
		return count == machineCount
	}, fmt.Sprintf("activeMachines to reach %d", machineCount))

	if lists := countActions(client, "list", "machines"); lists != 1 {
		t.Fatalf("expected 1 machine LIST for a burst of %d ADD events, got %d", machineCount, lists)
	}
}
