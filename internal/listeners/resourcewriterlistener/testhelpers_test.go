package resourcewriterlistener

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/vitistack-operator/pkg/consts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

const testVitistackName = "vitistack"

var kubernetesClusterGVR = schema.GroupVersionResource{
	Group:    "vitistack.io",
	Version:  "v1alpha1",
	Resource: "kubernetesclusters",
}

// useFakeDynamicClient swaps the shared dynamic client for a fake seeded with
// objs, and restores the original when the test finishes.
func useFakeDynamicClient(t *testing.T, objs ...runtime.Object) *dynamicfake.FakeDynamicClient {
	t.Helper()
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			vitistackGVR:         "VitistackList",
			machineGVR:           "MachineList",
			kubernetesClusterGVR: "KubernetesClusterList",
		},
		objs...,
	)
	previous := k8sclient.DynamicClient
	k8sclient.DynamicClient = client
	viper.Set(consts.VITISTACKCRDNAME, testVitistackName)
	t.Cleanup(func() { k8sclient.DynamicClient = previous })
	return client
}

func newTestVitistack() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "vitistack.io/v1alpha1",
		"kind":       "Vitistack",
		"metadata":   map[string]any{"name": testVitistackName},
	}}
}

// countActions returns how many API calls with the given verb were made against resource.
func countActions(client *dynamicfake.FakeDynamicClient, verb, resource string) int {
	n := 0
	for _, action := range client.Actions() {
		if action.GetVerb() == verb && action.GetResource().Resource == resource {
			n++
		}
	}
	return n
}

func getVitistackStatus(t *testing.T, client *dynamicfake.FakeDynamicClient) map[string]any {
	t.Helper()
	obj, err := client.Resource(vitistackGVR).Get(context.Background(), testVitistackName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get vitistack: %v", err)
	}
	status, _, _ := unstructured.NestedMap(obj.Object, "status")
	return status
}

// eventually polls cond until it returns true, failing the test after timeout.
func eventually(t *testing.T, timeout time.Duration, cond func() bool, waitingFor string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out after %s waiting for %s", timeout, waitingFor)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
