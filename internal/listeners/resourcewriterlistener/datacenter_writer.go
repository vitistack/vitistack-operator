package resourcewriterlistener

import (
	"context"
	"slices"
	"sync"

	"github.com/spf13/viper"
	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/vitistack-operator/internal/services/vitistacknameservice"
	"github.com/vitistack/vitistack-operator/pkg/consts"
	"github.com/vitistack/vitistack-operator/pkg/eventmanager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

// Provider type constants to determine which provider list to update
const (
	KubernetesProviderType = "kubernetesProviders"
	MachineProviderType    = "machineProviders"
)

func RegisterWriters() {
	eventmanager.EventBus.Subscribe("KubernetesProvider", handleKubernetesProviderEvents)
	eventmanager.EventBus.Subscribe("MachineProvider", handleMachineProviderEvents)
	eventmanager.EventBus.Subscribe("ConfigMap", handleConfigMapEvents)
}

// handleKubernetesProviderEvents processes events for KubernetesProvider resources
func handleKubernetesProviderEvents(event eventmanager.ResourceEvent) {
	updateVitistackWithProvider(event, KubernetesProviderType)
}

// handleMachineProviderEvents processes events for MachineProvider resources
func handleMachineProviderEvents(event eventmanager.ResourceEvent) {
	updateVitistackWithProvider(event, MachineProviderType)
}

func handleConfigMapEvents(event eventmanager.ResourceEvent) {
	vlog.Info("Received ConfigMap event",
		"type: ", string(event.Type),
		"name: ", event.Resource.GetName(),
		"namespace: ", event.Resource.GetNamespace())

	configMapName := viper.GetString(consts.CONFIGMAPNAME)
	if event.Resource.GetName() != configMapName {
		return
	}

	// Invalidate cache for this ConfigMap to ensure fresh data is used
	namespace := event.Resource.GetNamespace()
	err := vitistacknameservice.InvalidateCache(context.TODO(), namespace, configMapName)
	if err != nil {
		vlog.Error("Failed to invalidate ConfigMap cache", err)
	}

	updateVitistack(event)
}

func updateVitistack(event eventmanager.ResourceEvent) {
	if event.Resource.Object == nil {
		vlog.Error("Resource object is nil", nil)
		return
	}

	if event.Resource.GetName() != viper.GetString(consts.CONFIGMAPNAME) {
		return
	}

	if event.Resource.GetNamespace() != viper.GetString(consts.NAMESPACE) {
		return
	}

	vitistackCrdName := viper.GetString(consts.VITISTACKCRDNAME)

	// Extract ConfigMap data from the event
	configMapData, exists, err := unstructured.NestedStringMap(event.Resource.Object, "data")
	if err != nil {
		vlog.Error("Failed to extract ConfigMap data", err)
		return
	}
	if !exists || configMapData == nil {
		vlog.Error("ConfigMap data not found", nil)
		return
	}

	// Extract values from ConfigMap data
	vitistackName := configMapData["name"]
	region := configMapData["region"]
	zone := configMapData["zone"]
	country := configMapData["country"]
	provider := configMapData["provider"]
	description := configMapData["description"]
	infrastructure := configMapData["infrastructure"]

	vlog.Info("Processing ConfigMap update for Vitistack",
		"vitistackName: ", vitistackName,
		"region: ", region,
		"zone: ", zone,
		"country: ", country,
		"provider: ", provider,
		"infrastructure: ", infrastructure,
		"namespace: ", event.Resource.GetNamespace())

	if vitistackName == "" {
		vlog.Error("Vitistack name is empty in ConfigMap", nil)
		return
	}

	// Get or create the vitistack object
	vitistackRWMutex.RLock()
	vitistackObj, err := k8sclient.DynamicClient.Resource(vitistackGVR).Get(context.TODO(), vitistackCrdName, metav1.GetOptions{})
	vitistackRWMutex.RUnlock()
	if err != nil {
		// If the vitistack doesn't exist, we need to create it
		_, err = getOrCreateVitistackCrd(vitistackCrdName, "", "")
		if err != nil {
			vlog.Error("Failed to get or create Viti stack CRD", err,
				"name", event.Resource.GetName(),
				"namespace", event.Resource.GetNamespace())
			return
		}
		// Get the newly created vitistack object
		vitistackRWMutex.RLock()
		vitistackObj, err = k8sclient.DynamicClient.Resource(vitistackGVR).Get(context.TODO(), vitistackCrdName, metav1.GetOptions{})
		vitistackRWMutex.RUnlock()
		if err != nil {
			vlog.Error("Failed to get newly created Viti stack CRD", err)
			return
		}
	}

	// Update the fields in the vitistack object based on ConfigMap data
	updateNeeded := false

	// Update displayName
	if vitistackName != "" {
		err = unstructured.SetNestedField(vitistackObj.Object, vitistackName, "spec", "displayName")
		if err != nil {
			vlog.Error("Failed to set Viti stack displayName", err)
			return
		}
		updateNeeded = true
	}

	// Update region
	if region != "" {
		err = unstructured.SetNestedField(vitistackObj.Object, region, "spec", "region")
		if err != nil {
			vlog.Error("Failed to set Viti stack region", err)
			return
		}
		updateNeeded = true
	}

	// Update zone
	if zone != "" {
		err = unstructured.SetNestedField(vitistackObj.Object, zone, "spec", "zone")
		if err != nil {
			vlog.Error("Failed to set Viti stack zone", err)
			return
		}
		updateNeeded = true
	}

	// Update location (as a nested object with country field)
	if country != "" {
		locationObj := map[string]any{
			"country": country,
		}
		err = unstructured.SetNestedField(vitistackObj.Object, locationObj, "spec", "location")
		if err != nil {
			vlog.Error("Failed to set Viti stack country", err)
			return
		}
		updateNeeded = true
	}

	// Update description with provider information if available
	if description != "" {
		err = unstructured.SetNestedField(vitistackObj.Object, description, "spec", "description")
		if err != nil {
			vlog.Error("Failed to set Viti stack description", err)
			return
		}
		updateNeeded = true
	}

	if infrastructure != "" {
		err = unstructured.SetNestedField(vitistackObj.Object, infrastructure, "spec", "infrastructure")
		if err != nil {
			vlog.Error("Failed to set Viti stack infrastructure", err)
			return
		}
		updateNeeded = true
	}

	// Only update if changes were made
	if !updateNeeded {
		vlog.Info("No updates needed for Viti stack CRD")
		return
	}

	// Update the vitistack object
	_, err = k8sclient.DynamicClient.Resource(vitistackGVR).Update(context.TODO(), vitistackObj, metav1.UpdateOptions{})
	if err != nil {
		vlog.Error("Failed to update Viti stack CRD", err,
			"name", event.Resource.GetName(),
			"namespace", event.Resource.GetNamespace())
		return
	}

	vlog.Info("Updated Viti stack CRD from ConfigMap",
		"vitistackName: ", vitistackName,
		"region: ", region,
		"zone: ", zone,
		"country: ", country,
		"provider: ", provider,
		"namespace: ", event.Resource.GetNamespace())
}

// updateVitistackWithProvider handles updating the Viti stack CRD with provider information
// based on the providerType (either "kubernetesProviders" or "machineProviders")
func updateVitistackWithProvider(event eventmanager.ResourceEvent, providerType string) {
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
	vitistackObj, err := getOrCreateVitistackCrd(vitistackCrdName, providerName, providerType)
	if err != nil {
		vlog.Error("Failed to get or create Viti stack CRD", err,
			"name: ", vitistackCrdName)
		return
	}

	// Handle based on event type
	switch event.Type {
	case eventmanager.EventAdd, eventmanager.EventUpdate:
		addProviderToVitistack(vitistackObj, providerName, providerType)
	case eventmanager.EventDelete:
		removeProviderFromVitistack(vitistackObj, providerName, providerType)
	default:
		vlog.Info("Unhandled event type", "type: ", string(event.Type))
	}
}

// getOrCreateVitistackCrd tries to get an existing vitistack CRD or creates a new one if it doesn't exist
func getOrCreateVitistackCrd(name, providerName, providerType string) (*unstructured.Unstructured, error) {
	// Use a read lock first since we're just checking if the vitistack exists
	vitistackRWMutex.RLock()
	vitistackObj, err := k8sclient.DynamicClient.Resource(vitistackGVR).Get(context.TODO(), name, metav1.GetOptions{})
	vitistackRWMutex.RUnlock()

	if err == nil {
		return vitistackObj, nil
	}

	// Get vitistack information from ConfigMap
	vitistackName, region, country, zone := getVitistackInfoFromConfigMap()

	// If the vitistack doesn't exist, we need to create it - acquire a write lock
	vitistackRWMutex.Lock()
	defer vitistackRWMutex.Unlock()

	// Check again in case another goroutine created the vitistack while we were waiting for the lock
	vitistackObj, err = k8sclient.DynamicClient.Resource(vitistackGVR).Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil {
		return vitistackObj, nil
	}

	// Initialize with empty lists
	kubernetesProviders := []string{}
	machineProviders := []string{}

	// Add the current provider to the appropriate list
	switch providerType {
	case KubernetesProviderType:
		kubernetesProviders = append(kubernetesProviders, providerName)
	case MachineProviderType:
		machineProviders = append(machineProviders, providerName)
	}

	// Build location object if location is provided
	var locationObj map[string]any
	if country != "" {
		locationObj = map[string]any{
			"country": country,
		}
	}

	spec := map[string]any{
		"kubernetesProviders": kubernetesProviders,
		"machineProviders":    machineProviders,
		"displayName":         vitistackName,
	}

	// Add region if provided
	if region != "" {
		spec["region"] = region
	}

	// Add zone if provided
	if zone != "" {
		spec["zone"] = zone
	}

	// Add location if provided
	if locationObj != nil {
		spec["location"] = locationObj
	}

	vitistack := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "vitistack.io/v1alpha1",
			"kind":       "Vitistack",
			"metadata": map[string]any{
				"name": name,
			},
			"spec": spec,
		},
	}

	createdObj, err := k8sclient.DynamicClient.Resource(vitistackGVR).Create(context.TODO(), vitistack, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	vlog.Info("Created new Vitistack CRD",
		"name: ", name)

	return createdObj, nil
}

// getVitistackInfoFromConfigMap retrieves vitistack information from the ConfigMap
func getVitistackInfoFromConfigMap() (name, region, country, zone string) {
	// First try to get from cache/service
	ctx := context.TODO()
	vitistackName, err := vitistacknameservice.GetName(ctx)
	if err != nil {
		vlog.Error("Failed to get vitistack name from service", err)
		vitistackName = ""
	}

	// Try to get the ConfigMap directly to extract region, zone and location
	namespace := viper.GetString(consts.NAMESPACE)
	configMapName := viper.GetString(consts.CONFIGMAPNAME)

	if k8sclient.Kubernetes != nil {
		configMap, err := k8sclient.Kubernetes.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
		if err == nil && configMap.Data != nil {
			if vitistackName == "" {
				vitistackName = configMap.Data["name"]
			}
			region = configMap.Data["region"]
			country = configMap.Data["country"]
			zone = configMap.Data["zone"]
		} else {
			vlog.Error("Failed to get ConfigMap for vitistack info", err)
		}
	}

	return vitistackName, region, country, zone
}

// addProviderToVitistack adds a provider to the specified provider list if it doesn't already exist
func addProviderToVitistack(vitistackObj *unstructured.Unstructured, providerName, providerType string) {
	vitistackName := vitistackObj.GetName()

	// First, use a read lock to check if the provider already exists
	vitistackRWMutex.RLock()
	// Get the latest version of the vitistack object
	latestObj, err := k8sclient.DynamicClient.Resource(vitistackGVR).Get(context.TODO(), vitistackName, metav1.GetOptions{})
	if err != nil {
		vitistackRWMutex.RUnlock()
		vlog.Error("Failed to get Vitistack CRD", err,
			"name: ", vitistackName)
		return
	}

	// Check if provider already exists in the list
	providers, found, err := unstructured.NestedStringSlice(latestObj.Object, "spec", providerType)
	if err != nil {
		vitistackRWMutex.RUnlock()
		vlog.Error("Failed to get providers from vitistack", err,
			"providerType: ", providerType)
		return
	}

	if !found {
		providers = []string{}
	}

	providerExists := slices.Contains(providers, providerName)
	// If provider already exists, just log and return (no need for a write lock)
	if providerExists {
		vitistackRWMutex.RUnlock()
		vlog.Info("Provider already exists in Viti stack, no update needed",
			"providerType: ", providerType,
			"provider: ", providerName,
			"vitistack: ", vitistackName)
		return
	}

	// Release read lock before acquiring write lock to avoid deadlock
	vitistackRWMutex.RUnlock()

	// Acquire write lock for the update operation
	vitistackRWMutex.Lock()
	defer vitistackRWMutex.Unlock()

	// Get the latest version again after acquiring the write lock
	// This ensures we're working with current data even if it changed while we were waiting for the lock
	latestObj, err = k8sclient.DynamicClient.Resource(vitistackGVR).Get(context.TODO(), vitistackName, metav1.GetOptions{})
	if err != nil {
		vlog.Error("Failed to get updated Vitistack CRD", err,
			"name: ", vitistackName)
		return
	}

	// Re-check if provider exists (in case it was added while we were switching locks)
	providers, found, err = unstructured.NestedStringSlice(latestObj.Object, "spec", providerType)
	if err != nil {
		vlog.Error("Failed to get providers from vitistack", err,
			"providerType: ", providerType)
		return
	}

	if !found {
		providers = []string{}
	}

	providerExists = slices.Contains(providers, providerName)
	if !providerExists {
		// Add the provider to the list
		providers = append(providers, providerName)

		// Update the unstructured object
		err = unstructured.SetNestedStringSlice(latestObj.Object, providers, "spec", providerType)
		if err != nil {
			vlog.Error("Failed to set providers in vitistack", err,
				"providerType: ", providerType)
			return
		}

		// Update the vitistack resource
		_, err = k8sclient.DynamicClient.Resource(vitistackGVR).Update(context.TODO(), latestObj, metav1.UpdateOptions{})
		if err != nil {
			vlog.Error("Failed to update Viti stack CRD", err,
				"name: ", vitistackName)
			return
		}

		vlog.Info("Updated Viti stack CRD with provider",
			"name: ", vitistackName,
			"providerType: ", providerType,
			"provider: ", providerName)
	} else {
		vlog.Info("Provider already exists in Viti stack, no update needed",
			"providerType: ", providerType,
			"provider: ", providerName,
			"vitistack: ", vitistackName)
	}
}

// removeProviderFromVitistack removes a provider from the specified provider list
func removeProviderFromVitistack(vitistackObj *unstructured.Unstructured, providerName, providerType string) {
	vitistackName := vitistackObj.GetName()

	// First, use a read lock to check if the provider exists
	vitistackRWMutex.RLock()
	// Get the latest version of the vitistack object
	latestObj, err := k8sclient.DynamicClient.Resource(vitistackGVR).Get(context.TODO(), vitistackName, metav1.GetOptions{})
	if err != nil {
		vitistackRWMutex.RUnlock()
		vlog.Error("Failed to get Viti stack CRD", err,
			"name: ", vitistackName)
		return
	}

	// Check the current providers list
	providers, found, err := unstructured.NestedStringSlice(latestObj.Object, "spec", providerType)
	if err != nil {
		vitistackRWMutex.RUnlock()
		vlog.Error("Failed to get providers from vitistack", err,
			"providerType: ", providerType)
		return
	}

	if !found || len(providers) == 0 {
		// Nothing to remove
		vitistackRWMutex.RUnlock()
		vlog.Info("No providers found in vitistack to remove",
			"providerType: ", providerType,
			"vitistack: ", vitistackName)
		return
	}

	// Check if provider exists
	providerIndex := -1
	for i, provider := range providers {
		if provider == providerName {
			providerIndex = i
			break
		}
	}

	// If provider doesn't exist, just log and return (no need for a write lock)
	if providerIndex < 0 {
		vitistackRWMutex.RUnlock()
		vlog.Info("Provider not found in Viti stack, no removal needed",
			"providerType: ", providerType,
			"provider: ", providerName,
			"vitistack: ", vitistackName)
		return
	}

	// Release read lock before acquiring write lock to avoid deadlock
	vitistackRWMutex.RUnlock()

	// Acquire write lock for the update operation
	vitistackRWMutex.Lock()
	defer vitistackRWMutex.Unlock()

	// Get the latest version again after acquiring the write lock
	latestObj, err = k8sclient.DynamicClient.Resource(vitistackGVR).Get(context.TODO(), vitistackName, metav1.GetOptions{})
	if err != nil {
		vlog.Error("Failed to get updated Viti stack CRD", err,
			"name: ", vitistackName)
		return
	}

	// Re-check the providers (in case they changed while we were switching locks)
	providers, found, err = unstructured.NestedStringSlice(latestObj.Object, "spec", providerType)
	if err != nil {
		vlog.Error("Failed to get providers from vitistack", err,
			"providerType: ", providerType)
		return
	}

	if !found || len(providers) == 0 {
		vlog.Info("No providers found in vitistack to remove",
			"providerType: ", providerType,
			"vitistack: ", vitistackName)
		return
	}

	// Re-check if provider exists and find its index
	providerIndex = -1
	for i, provider := range providers {
		if provider == providerName {
			providerIndex = i
			break
		}
	}

	if providerIndex >= 0 {
		// Remove the provider from the list
		providers = slices.Delete(providers, providerIndex, providerIndex+1)

		// Update the unstructured object
		err = unstructured.SetNestedStringSlice(latestObj.Object, providers, "spec", providerType)
		if err != nil {
			vlog.Error("Failed to update providers in vitistack", err,
				"providerType: ", providerType)
			return
		}

		// Update the vitistack resource
		_, err = k8sclient.DynamicClient.Resource(vitistackGVR).Update(context.TODO(), latestObj, metav1.UpdateOptions{})
		if err != nil {
			vlog.Error("Failed to update Viti stack CRD after removal", err,
				"name: ", vitistackName)
			return
		}

		vlog.Info("Removed provider from Viti stack CRD",
			"name: ", vitistackName,
			"providerType: ", providerType,
			"provider: ", providerName)
	} else {
		vlog.Info("Provider not found in Viti stack, no removal needed",
			"providerType: ", providerType,
			"provider: ", providerName,
			"vitistack: ", vitistackName)
	}
}
