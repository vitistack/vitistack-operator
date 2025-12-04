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

// handleKubernetesProviderEvents processes events for KubernetesProvider resources
func handleKubernetesProviderEvents(event eventmanager.ResourceEvent) {
	if event.Resource == nil {
		vlog.Error("Resource is nil in KubernetesProvider event", nil)
		return
	}
	updateVitistackStatusWithProvider(event, KubernetesProviderType)
}

// handleMachineProviderEvents processes events for MachineProvider resources
func handleMachineProviderEvents(event eventmanager.ResourceEvent) {
	if event.Resource == nil {
		vlog.Error("Resource is nil in MachineProvider event", nil)
		return
	}
	updateVitistackStatusWithProvider(event, MachineProviderType)
}

// updateVitistackStatusWithProvider handles updating the Viti stack CRD status with provider information
// based on the providerType (either "kubernetesProviders" or "machineProviders")
func updateVitistackStatusWithProvider(event eventmanager.ResourceEvent, providerType string) {
	// Use the shared dynamic client
	if k8sclient.DynamicClient == nil {
		vlog.Error("Dynamic client is not initialized", nil)
		return
	}

	providerName := event.Resource.GetName()
	namespace := event.Resource.GetNamespace()

	vlog.Info("Processing provider event",
		"type: ", string(event.Type),
		"providerType: ", providerType,
		"name: ", providerName,
		"namespace: ", namespace)

	// Get or create the vitistack CRD
	vitistackCrdName := viper.GetString(consts.VITISTACKCRDNAME)
	vitistackObj, err := getOrCreateVitistackCrd(vitistackCrdName)
	if err != nil {
		vlog.Error("Failed to get or create Viti stack CRD", err,
			"name: ", vitistackCrdName)
		return
	}

	// Extract provider metadata from the event resource
	providerMetadata := extractProviderMetadata(event)

	// Handle based on event type
	switch event.Type {
	case eventmanager.EventAdd, eventmanager.EventUpdate:
		addProviderToVitistackStatus(vitistackObj, providerName, providerType, providerMetadata)
	case eventmanager.EventDelete:
		removeProviderFromVitistackStatus(vitistackObj, providerName, providerType)
	default:
		vlog.Info("Unhandled event type", "type: ", string(event.Type))
	}
}

// extractProviderMetadata extracts metadata from the provider resource for status
func extractProviderMetadata(event eventmanager.ResourceEvent) map[string]any {
	metadata := map[string]any{
		"name":         event.Resource.GetName(),
		"namespace":    event.Resource.GetNamespace(),
		"discoveredAt": time.Now().UTC().Format(time.RFC3339),
	}

	// Try to extract providerType from spec
	if providerType, found, err := unstructured.NestedString(event.Resource.Object, "spec", "providerType"); err == nil && found {
		metadata["providerType"] = providerType
	}

	// Try to extract region from spec
	if region, found, err := unstructured.NestedString(event.Resource.Object, "spec", "region"); err == nil && found {
		metadata["region"] = region
	}

	// Try to extract zone from spec
	if zone, found, err := unstructured.NestedString(event.Resource.Object, "spec", "zone"); err == nil && found {
		metadata["zone"] = zone
	}

	// Try to extract ready status
	if phase, found, err := unstructured.NestedString(event.Resource.Object, "status", "phase"); err == nil && found {
		metadata["ready"] = phase == "Ready"
	}

	return metadata
}

// addProviderToVitistackStatus adds a provider to the status provider list if it doesn't already exist
func addProviderToVitistackStatus(vitistackObj *unstructured.Unstructured, providerName, providerType string, metadata map[string]any) {
	vitistackName := vitistackObj.GetName()

	// Acquire write lock for the update operation
	vitistackRWMutex.Lock()
	defer vitistackRWMutex.Unlock()

	// Get the latest version of the vitistack object
	latestObj, err := k8sclient.DynamicClient.Resource(vitistackGVR).Get(context.TODO(), vitistackName, metav1.GetOptions{})
	if err != nil {
		vlog.Error("Failed to get Vitistack CRD", err,
			"name: ", vitistackName)
		return
	}

	// Check if provider already exists in the status list
	providers, found, err := unstructured.NestedSlice(latestObj.Object, "status", providerType)
	if err != nil {
		vlog.Error("Failed to get providers from vitistack status", err,
			"providerType: ", providerType)
		return
	}

	if !found {
		providers = []any{}
	}

	// Check if provider already exists by name
	providerExists := false
	providerIndex := -1
	for i, p := range providers {
		if pMap, ok := p.(map[string]any); ok {
			if pMap["name"] == providerName {
				providerExists = true
				providerIndex = i
				break
			}
		}
	}

	if providerExists {
		// Update existing provider with new metadata
		providers[providerIndex] = metadata
	} else {
		// Add the new provider
		providers = append(providers, metadata)
	}

	// Ensure status exists
	status, _, _ := unstructured.NestedMap(latestObj.Object, "status")
	if status == nil {
		status = map[string]any{}
	}

	// Update the providers list in status
	status[providerType] = providers

	// Update the provider count
	countField := getProviderCountField(providerType)
	if countField != "" {
		status[countField] = int64(len(providers))
	}

	// Set the updated status
	err = unstructured.SetNestedField(latestObj.Object, status, "status")
	if err != nil {
		vlog.Error("Failed to set status in vitistack", err)
		return
	}

	// Update the vitistack resource status using UpdateStatus
	_, err = k8sclient.DynamicClient.Resource(vitistackGVR).UpdateStatus(context.TODO(), latestObj, metav1.UpdateOptions{})
	if err != nil {
		vlog.Error("Failed to update Viti stack CRD status", err,
			"name: ", vitistackName)
		return
	}

	if providerExists {
		vlog.Info("Updated provider in Viti stack status",
			"name: ", vitistackName,
			"providerType: ", providerType,
			"provider: ", providerName)
	} else {
		vlog.Info("Added provider to Viti stack status",
			"name: ", vitistackName,
			"providerType: ", providerType,
			"provider: ", providerName)
	}
}

// removeProviderFromVitistackStatus removes a provider from the status provider list
func removeProviderFromVitistackStatus(vitistackObj *unstructured.Unstructured, providerName, providerType string) {
	vitistackName := vitistackObj.GetName()

	// Acquire write lock for the update operation
	vitistackRWMutex.Lock()
	defer vitistackRWMutex.Unlock()

	// Get the latest version of the vitistack object
	latestObj, err := k8sclient.DynamicClient.Resource(vitistackGVR).Get(context.TODO(), vitistackName, metav1.GetOptions{})
	if err != nil {
		vlog.Error("Failed to get Viti stack CRD", err,
			"name: ", vitistackName)
		return
	}

	// Check the current providers list in status
	providers, found, err := unstructured.NestedSlice(latestObj.Object, "status", providerType)
	if err != nil {
		vlog.Error("Failed to get providers from vitistack status", err,
			"providerType: ", providerType)
		return
	}

	if !found || len(providers) == 0 {
		vlog.Info("No providers found in vitistack status to remove",
			"providerType: ", providerType,
			"vitistack: ", vitistackName)
		return
	}

	// Find provider index by name
	providerIndex := -1
	for i, p := range providers {
		if pMap, ok := p.(map[string]any); ok {
			if pMap["name"] == providerName {
				providerIndex = i
				break
			}
		}
	}

	if providerIndex < 0 {
		vlog.Info("Provider not found in Viti stack status, no removal needed",
			"providerType: ", providerType,
			"provider: ", providerName,
			"vitistack: ", vitistackName)
		return
	}

	// Remove the provider from the list
	providers = slices.Delete(providers, providerIndex, providerIndex+1)

	// Ensure status exists
	status, _, _ := unstructured.NestedMap(latestObj.Object, "status")
	if status == nil {
		status = map[string]any{}
	}

	// Update the providers list in status
	status[providerType] = providers

	// Update the provider count
	countField := getProviderCountField(providerType)
	if countField != "" {
		status[countField] = int64(len(providers))
	}

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

	vlog.Info("Removed provider from Viti stack status",
		"name: ", vitistackName,
		"providerType: ", providerType,
		"provider: ", providerName)
}
