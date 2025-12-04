package resourcewriterlistener

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupVersionResource for Vitistack CRD
var vitistackGVR = schema.GroupVersionResource{
	Group:    "vitistack.io",
	Version:  "v1alpha1",
	Resource: "vitistacks",
}

// Package-level mutex for vitistack operations to prevent race conditions
var vitistackRWMutex = &sync.RWMutex{}

// Provider type constants to determine which provider list to update in status
const (
	KubernetesProviderType = "kubernetesProviders"
	MachineProviderType    = "machineProviders"
	MachineClassType       = "machineClasses"
	ClusterType            = "clusters"
)

// getProviderCountField returns the count field name for a provider type
func getProviderCountField(providerType string) string {
	switch providerType {
	case KubernetesProviderType:
		return "kubernetesProviderCount"
	case MachineProviderType:
		return "machineProviderCount"
	default:
		return ""
	}
}

// initializeStatusDefaults ensures all required status fields have default values
func initializeStatusDefaults(status map[string]any) {
	// Initialize empty lists if not present
	if _, exists := status["kubernetesProviders"]; !exists {
		status["kubernetesProviders"] = []interface{}{}
	}
	if _, exists := status["machineProviders"]; !exists {
		status["machineProviders"] = []interface{}{}
	}
	if _, exists := status["machineClasses"]; !exists {
		status["machineClasses"] = []interface{}{}
	}
	if _, exists := status["clusters"]; !exists {
		status["clusters"] = []interface{}{}
	}
	if _, exists := status["providerStatuses"]; !exists {
		status["providerStatuses"] = []interface{}{}
	}
	if _, exists := status["conditions"]; !exists {
		status["conditions"] = []interface{}{}
	}

	// Initialize counts to 0 if not present
	if _, exists := status["kubernetesProviderCount"]; !exists {
		status["kubernetesProviderCount"] = int64(0)
	}
	if _, exists := status["machineProviderCount"]; !exists {
		status["machineProviderCount"] = int64(0)
	}
	if _, exists := status["activeClusters"]; !exists {
		status["activeClusters"] = int64(0)
	}
	if _, exists := status["activeMachines"]; !exists {
		status["activeMachines"] = int64(0)
	}

	// Initialize phase if not present
	if _, exists := status["phase"]; !exists {
		status["phase"] = "Ready"
	}
}
