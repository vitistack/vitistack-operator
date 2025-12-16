package resourcewriterlistener

import (
	"context"
	"slices"
	"time"

	"github.com/spf13/viper"
	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/vitistack-operator/pkg/consts"
	"github.com/vitistack/vitistack-operator/pkg/eventmanager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// handleKubernetesClusterEvents processes events for KubernetesCluster resources
func handleKubernetesClusterEvents(event eventmanager.ResourceEvent) {
	if event.Resource == nil {
		vlog.Error("Resource is nil in KubernetesCluster event", nil)
		return
	}
	updateVitistackStatusWithCluster(event)
}

// updateVitistackStatusWithCluster handles updating the Viti stack CRD status with cluster information
func updateVitistackStatusWithCluster(event eventmanager.ResourceEvent) {
	// Use the shared dynamic client
	if k8sclient.DynamicClient == nil {
		vlog.Error("Dynamic client is not initialized", nil)
		return
	}

	clusterName := event.Resource.GetName()

	// Get or create the vitistack CRD
	vitistackCrdName := viper.GetString(consts.VITISTACKCRDNAME)
	vitistackObj, err := getOrCreateVitistackCrd(vitistackCrdName)
	if err != nil {
		vlog.Error("Failed to get or create Viti stack CRD", err,
			"name: ", vitistackCrdName)
		return
	}

	// Extract cluster metadata from the event resource
	clusterMetadata := extractClusterMetadata(event)

	// Handle based on event type
	switch event.Type {
	case eventmanager.EventAdd, eventmanager.EventUpdate:
		addClusterToVitistackStatus(vitistackObj, clusterName, clusterMetadata)
	case eventmanager.EventDelete:
		removeClusterFromVitistackStatus(vitistackObj, clusterName)
	}
}

// extractClusterMetadata extracts metadata from the cluster resource for status
func extractClusterMetadata(event eventmanager.ResourceEvent) map[string]any {
	metadata := map[string]any{
		"name":         event.Resource.GetName(),
		"namespace":    event.Resource.GetNamespace(),
		"discoveredAt": time.Now().UTC().Format(time.RFC3339),
	}

	// Try to extract version from spec
	if version, found, err := unstructured.NestedString(event.Resource.Object, "spec", "version"); err == nil && found {
		metadata["version"] = version
	}

	// Try to extract phase from status
	if phase, found, err := unstructured.NestedString(event.Resource.Object, "status", "phase"); err == nil && found {
		metadata["phase"] = phase
		metadata["ready"] = phase == "Running"
	}

	// Try to extract control plane replicas from spec.topology.controlplane.replicas
	if replicas, found, err := unstructured.NestedInt64(event.Resource.Object, "spec", "topology", "controlplane", "replicas"); err == nil && found {
		metadata["controlPlaneReplicas"] = replicas
	}

	// Try to extract worker replicas - sum from spec.topology.workers
	workers, found, err := unstructured.NestedSlice(event.Resource.Object, "spec", "topology", "workers")
	if err == nil && found {
		var totalWorkerReplicas int64
		for _, worker := range workers {
			if workerMap, ok := worker.(map[string]any); ok {
				if replicas, ok := workerMap["replicas"].(int64); ok {
					totalWorkerReplicas += replicas
				}
			}
		}
		metadata["workerReplicas"] = totalWorkerReplicas
	}

	return metadata
}

// addClusterToVitistackStatus adds a cluster to the status clusters list
func addClusterToVitistackStatus(vitistackObj *unstructured.Unstructured, clusterName string, metadata map[string]any) {
	vitistackName := vitistackObj.GetName()

	// Acquire write lock for the update operation
	vitistackRWMutex.Lock()
	defer vitistackRWMutex.Unlock()

	// Get the latest version
	latestObj, err := k8sclient.DynamicClient.Resource(vitistackGVR).Get(context.TODO(), vitistackName, metav1.GetOptions{})
	if err != nil {
		vlog.Error("Failed to get Vitistack CRD", err,
			"name: ", vitistackName)
		return
	}

	// Get current clusters from status
	clusters, found, err := unstructured.NestedSlice(latestObj.Object, "status", "clusters")
	if err != nil {
		vlog.Error("Failed to get clusters from vitistack status", err)
		return
	}

	if !found {
		clusters = []any{}
	}

	// Check if cluster already exists and update or add
	clusterExists := false
	clusterIndex := -1
	for i, c := range clusters {
		if cMap, ok := c.(map[string]any); ok {
			if cMap["name"] == clusterName {
				clusterExists = true
				clusterIndex = i
				break
			}
		}
	}

	if clusterExists {
		// Check if metadata has actually changed (skip discoveredAt comparison)
		existingCluster, ok := clusters[clusterIndex].(map[string]any)
		if ok && clusterMetadataEqual(existingCluster, metadata) {
			// No changes, skip update silently
			return
		}
		// Preserve original discoveredAt timestamp
		if existingDiscoveredAt, ok := existingCluster["discoveredAt"]; ok {
			metadata["discoveredAt"] = existingDiscoveredAt
		}
		// Update existing cluster
		clusters[clusterIndex] = metadata
	} else {
		// Add new cluster
		clusters = append(clusters, metadata)
	}

	// Ensure status exists
	status, _, _ := unstructured.NestedMap(latestObj.Object, "status")
	if status == nil {
		status = map[string]any{}
	}

	// Update clusters in status
	status["clusters"] = clusters

	// Update activeClusters count
	status["activeClusters"] = int64(len(clusters))

	// Set the updated status
	err = unstructured.SetNestedField(latestObj.Object, status, "status")
	if err != nil {
		vlog.Error("Failed to set status in vitistack", err)
		return
	}

	// Update the vitistack resource status
	_, err = k8sclient.DynamicClient.Resource(vitistackGVR).UpdateStatus(context.TODO(), latestObj, metav1.UpdateOptions{})
	if err != nil {
		vlog.Error("Failed to update Viti stack CRD status", err,
			"name: ", vitistackName)
		return
	}

	if !clusterExists {
		vlog.Info("Added cluster to Viti stack status",
			" name: ", vitistackName,
			" cluster: ", clusterName)
	}
}

// removeClusterFromVitistackStatus removes a cluster from the status clusters list
func removeClusterFromVitistackStatus(vitistackObj *unstructured.Unstructured, clusterName string) {
	vitistackName := vitistackObj.GetName()

	// Acquire write lock for the update operation
	vitistackRWMutex.Lock()
	defer vitistackRWMutex.Unlock()

	// Get the latest version
	latestObj, err := k8sclient.DynamicClient.Resource(vitistackGVR).Get(context.TODO(), vitistackName, metav1.GetOptions{})
	if err != nil {
		vlog.Error("Failed to get Viti stack CRD", err,
			"name: ", vitistackName)
		return
	}

	// Get current clusters from status
	clusters, found, err := unstructured.NestedSlice(latestObj.Object, "status", "clusters")
	if err != nil {
		vlog.Error("Failed to get clusters from vitistack status", err)
		return
	}

	if !found || len(clusters) == 0 {
		vlog.Info("No clusters found in vitistack status to remove",
			"vitistack: ", vitistackName)
		return
	}

	// Find and remove the cluster
	clusterIndex := -1
	for i, c := range clusters {
		if cMap, ok := c.(map[string]any); ok {
			if cMap["name"] == clusterName {
				clusterIndex = i
				break
			}
		}
	}

	if clusterIndex < 0 {
		vlog.Info("Cluster not found in Viti stack status, no removal needed",
			"cluster: ", clusterName,
			"vitistack: ", vitistackName)
		return
	}

	// Remove the cluster
	clusters = slices.Delete(clusters, clusterIndex, clusterIndex+1)

	// Ensure status exists
	status, _, _ := unstructured.NestedMap(latestObj.Object, "status")
	if status == nil {
		status = map[string]any{}
	}

	// Update clusters in status
	status["clusters"] = clusters

	// Update activeClusters count
	status["activeClusters"] = int64(len(clusters))

	// Set the updated status
	err = unstructured.SetNestedField(latestObj.Object, status, "status")
	if err != nil {
		vlog.Error("Failed to set status in vitistack", err)
		return
	}

	// Update the vitistack resource status
	_, err = k8sclient.DynamicClient.Resource(vitistackGVR).UpdateStatus(context.TODO(), latestObj, metav1.UpdateOptions{})
	if err != nil {
		vlog.Error("Failed to update Viti stack CRD status after removal", err,
			"name: ", vitistackName)
		return
	}

	vlog.Info("Removed cluster from Viti stack status",
		"name: ", vitistackName,
		"cluster: ", clusterName)
}
