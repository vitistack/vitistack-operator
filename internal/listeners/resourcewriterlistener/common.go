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

// GroupVersionResource for Machine CRD
var machineGVR = schema.GroupVersionResource{
	Group:    "vitistack.io",
	Version:  "v1alpha1",
	Resource: "machines",
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
		status["kubernetesProviders"] = []any{}
	}
	if _, exists := status["machineProviders"]; !exists {
		status["machineProviders"] = []any{}
	}
	if _, exists := status["machineClasses"]; !exists {
		status["machineClasses"] = []any{}
	}
	if _, exists := status["clusters"]; !exists {
		status["clusters"] = []any{}
	}
	if _, exists := status["providerStatuses"]; !exists {
		status["providerStatuses"] = []any{}
	}
	if _, exists := status["conditions"]; !exists {
		status["conditions"] = []any{}
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

// clusterMetadataEqual compares two cluster metadata maps to check if they're equal
// Ignores discoveredAt field since it changes on each event
func clusterMetadataEqual(existing, new map[string]any) bool {
	// Fields to compare (skip discoveredAt)
	fieldsToCompare := []string{"name", "namespace", "version", "phase", "ready", "controlPlaneReplicas", "workerReplicas"}

	for _, field := range fieldsToCompare {
		existingVal, existingOk := existing[field]
		newVal, newOk := new[field]

		// If one has the field and the other doesn't, they're different
		if existingOk != newOk {
			return false
		}

		// If both have the field, compare values
		if existingOk && newOk {
			// Handle type differences (int64 vs float64 from JSON)
			switch ev := existingVal.(type) {
			case int64:
				switch nv := newVal.(type) {
				case int64:
					if ev != nv {
						return false
					}
				case float64:
					if ev != int64(nv) {
						return false
					}
				default:
					return false
				}
			case float64:
				switch nv := newVal.(type) {
				case int64:
					if int64(ev) != nv {
						return false
					}
				case float64:
					if ev != nv {
						return false
					}
				default:
					return false
				}
			case bool:
				if nv, ok := newVal.(bool); !ok || ev != nv {
					return false
				}
			case string:
				if nv, ok := newVal.(string); !ok || ev != nv {
					return false
				}
			default:
				// For other types, use simple equality
				if existingVal != newVal {
					return false
				}
			}
		}
	}

	return true
}

// providerMetadataEqual compares two provider metadata maps to check if they're equal
func providerMetadataEqual(existing, new map[string]any) bool {
	// Fields to compare (skip discoveredAt)
	fieldsToCompare := []string{"name", "namespace", "type", "ready"}

	for _, field := range fieldsToCompare {
		existingVal, existingOk := existing[field]
		newVal, newOk := new[field]

		if existingOk != newOk {
			return false
		}

		if existingOk && newOk && existingVal != newVal {
			return false
		}
	}

	return true
}
