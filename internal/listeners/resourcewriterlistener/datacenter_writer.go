package resourcewriterlistener

import (
	"context"
	"slices"
	"sync"

	"github.com/NorskHelsenett/ror/pkg/rlog"
	"github.com/spf13/viper"
	"github.com/vitistack/datacenter-operator/internal/clients"
	"github.com/vitistack/datacenter-operator/internal/services/datacenternameservice"
	"github.com/vitistack/datacenter-operator/pkg/consts"
	"github.com/vitistack/datacenter-operator/pkg/eventmanager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupVersionResource for Datacenter CRD
var datacenterGVR = schema.GroupVersionResource{
	Group:    "vitistack.io",
	Version:  "v1alpha1",
	Resource: "datacenters",
}

// Package-level mutex for datacenter operations to prevent race conditions
var datacenterRWMutex = &sync.RWMutex{}

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
	updateDatacenterWithProvider(event, KubernetesProviderType)
}

// handleMachineProviderEvents processes events for MachineProvider resources
func handleMachineProviderEvents(event eventmanager.ResourceEvent) {
	updateDatacenterWithProvider(event, MachineProviderType)
}

func handleConfigMapEvents(event eventmanager.ResourceEvent) {
	rlog.Info("Received ConfigMap event",
		rlog.String("type", string(event.Type)),
		rlog.String("name", event.Resource.GetName()),
		rlog.String("namespace", event.Resource.GetNamespace()))

	configMapName := viper.GetString(consts.CONFIGMAPNAME)
	if event.Resource.GetName() != configMapName {
		return
	}

	// Invalidate cache for this ConfigMap to ensure fresh data is used
	namespace := event.Resource.GetNamespace()
	err := datacenternameservice.InvalidateCache(context.TODO(), namespace, configMapName)
	if err != nil {
		rlog.Error("Failed to invalidate ConfigMap cache", err)
	}

	updateDatacenter(event)
}

func updateDatacenter(event eventmanager.ResourceEvent) {
	if event.Resource.Object == nil {
		rlog.Error("Resource object is nil", nil)
		return
	}

	if event.Resource.GetName() != viper.GetString(consts.CONFIGMAPNAME) {
		return
	}

	if event.Resource.GetNamespace() != viper.GetString(consts.NAMESPACE) {
		return
	}

	datacenterCrdName := viper.GetString(consts.DATACENTERCRDNAME)

	// Extract ConfigMap data from the event
	configMapData, exists, err := unstructured.NestedStringMap(event.Resource.Object, "data")
	if err != nil {
		rlog.Error("Failed to extract ConfigMap data", err)
		return
	}
	if !exists || configMapData == nil {
		rlog.Error("ConfigMap data not found", nil)
		return
	}

	// Extract values from ConfigMap data
	datacenterName := configMapData["name"]
	region := configMapData["region"]
	zone := configMapData["zone"]
	location := configMapData["location"]
	provider := configMapData["provider"]
	description := configMapData["description"]

	rlog.Info("Processing ConfigMap update for Datacenter",
		rlog.String("datacenterName", datacenterName),
		rlog.String("region", region),
		rlog.String("zone", zone),
		rlog.String("location", location),
		rlog.String("provider", provider))

	if datacenterName == "" {
		rlog.Error("Datacenter name is empty in ConfigMap", nil)
		return
	}

	// Get or create the datacenter object
	datacenterRWMutex.RLock()
	datacenterObj, err := clients.DynamicClient.Resource(datacenterGVR).Get(context.TODO(), datacenterCrdName, metav1.GetOptions{})
	datacenterRWMutex.RUnlock()
	if err != nil {
		// If the datacenter doesn't exist, we need to create it
		_, err = getOrCreateDatacenterCrd(datacenterName, "", "")
		if err != nil {
			rlog.Error("Failed to get or create Datacenter CRD", err,
				rlog.String("name", event.Resource.GetName()),
				rlog.String("namespace", event.Resource.GetNamespace()))
			return
		}
		// Get the newly created datacenter object
		datacenterRWMutex.RLock()
		datacenterObj, err = clients.DynamicClient.Resource(datacenterGVR).Get(context.TODO(), datacenterCrdName, metav1.GetOptions{})
		datacenterRWMutex.RUnlock()
		if err != nil {
			rlog.Error("Failed to get newly created Datacenter CRD", err)
			return
		}
	}

	// Update the fields in the datacenter object based on ConfigMap data
	updateNeeded := false

	// Update displayName
	if datacenterName != "" {
		err = unstructured.SetNestedField(datacenterObj.Object, datacenterName, "spec", "displayName")
		if err != nil {
			rlog.Error("Failed to set Datacenter displayName", err)
			return
		}
		updateNeeded = true
	}

	// Update region
	if region != "" {
		err = unstructured.SetNestedField(datacenterObj.Object, region, "spec", "region")
		if err != nil {
			rlog.Error("Failed to set Datacenter region", err)
			return
		}
		updateNeeded = true
	}

	// Update zone
	if zone != "" {
		err = unstructured.SetNestedField(datacenterObj.Object, zone, "spec", "zone")
		if err != nil {
			rlog.Error("Failed to set Datacenter zone", err)
			return
		}
		updateNeeded = true
	}

	// Update location (as a nested object with country field)
	if location != "" {
		locationObj := map[string]any{
			"country": location,
		}
		err = unstructured.SetNestedField(datacenterObj.Object, locationObj, "spec", "location")
		if err != nil {
			rlog.Error("Failed to set Datacenter location", err)
			return
		}
		updateNeeded = true
	}

	// Update description with provider information if available
	if description != "" {
		err = unstructured.SetNestedField(datacenterObj.Object, description, "spec", "description")
		if err != nil {
			rlog.Error("Failed to set Datacenter description", err)
			return
		}
		updateNeeded = true
	}

	// Only update if changes were made
	if !updateNeeded {
		rlog.Info("No updates needed for Datacenter CRD")
		return
	}

	// Update the datacenter object
	_, err = clients.DynamicClient.Resource(datacenterGVR).Update(context.TODO(), datacenterObj, metav1.UpdateOptions{})
	if err != nil {
		rlog.Error("Failed to update Datacenter CRD", err,
			rlog.String("name", event.Resource.GetName()),
			rlog.String("namespace", event.Resource.GetNamespace()))
		return
	}

	rlog.Info("Updated Datacenter CRD from ConfigMap",
		rlog.String("datacenterName", datacenterName),
		rlog.String("region", region),
		rlog.String("zone", zone),
		rlog.String("location", location),
		rlog.String("provider", provider),
		rlog.String("namespace", event.Resource.GetNamespace()))
}

// updateDatacenterWithProvider handles updating the Datacenter CRD with provider information
// based on the providerType (either "kubernetesProviders" or "machineProviders")
func updateDatacenterWithProvider(event eventmanager.ResourceEvent, providerType string) {
	// Use the shared dynamic client
	if clients.DynamicClient == nil {
		rlog.Error("Dynamic client is not initialized", nil)
		return
	}

	providerName := event.Resource.GetName()
	namespace := event.Resource.GetNamespace()

	rlog.Info("Processing provider event",
		rlog.String("type", string(event.Type)),
		rlog.String("providerType", providerType),
		rlog.String("name", providerName),
		rlog.String("namespace", namespace))

	// Get or create the datacenter CRD
	datacenterCrdName := viper.GetString(consts.DATACENTERCRDNAME)
	datacenterObj, err := getOrCreateDatacenterCrd(datacenterCrdName, providerName, providerType)
	if err != nil {
		rlog.Error("Failed to get or create Datacenter CRD", err,
			rlog.String("name", datacenterCrdName))
		return
	}

	// Handle based on event type
	switch event.Type {
	case eventmanager.EventAdd, eventmanager.EventUpdate:
		addProviderToDatacenter(datacenterObj, providerName, providerType)
	case eventmanager.EventDelete:
		removeProviderFromDatacenter(datacenterObj, providerName, providerType)
	default:
		rlog.Info("Unhandled event type", rlog.String("type", string(event.Type)))
	}
}

// getOrCreateDatacenterCrd tries to get an existing datacenter CRD or creates a new one if it doesn't exist
func getOrCreateDatacenterCrd(name, providerName, providerType string) (*unstructured.Unstructured, error) {
	// Use a read lock first since we're just checking if the datacenter exists
	datacenterRWMutex.RLock()
	datacenterObj, err := clients.DynamicClient.Resource(datacenterGVR).Get(context.TODO(), name, metav1.GetOptions{})
	datacenterRWMutex.RUnlock()

	if err == nil {
		return datacenterObj, nil
	}

	// Get datacenter information from ConfigMap
	datacenterName, region, location, zone := getDatacenterInfoFromConfigMap()

	// If the datacenter doesn't exist, we need to create it - acquire a write lock
	datacenterRWMutex.Lock()
	defer datacenterRWMutex.Unlock()

	// Check again in case another goroutine created the datacenter while we were waiting for the lock
	datacenterObj, err = clients.DynamicClient.Resource(datacenterGVR).Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil {
		return datacenterObj, nil
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
	if location != "" {
		locationObj = map[string]any{
			"country": location,
		}
	}

	spec := map[string]any{
		"kubernetesProviders": kubernetesProviders,
		"machineProviders":    machineProviders,
		"displayName":         datacenterName,
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

	datacenter := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "vitistack.io/v1alpha1",
			"kind":       "Datacenter",
			"metadata": map[string]any{
				"name": name,
			},
			"spec": spec,
		},
	}

	createdObj, err := clients.DynamicClient.Resource(datacenterGVR).Create(context.TODO(), datacenter, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	rlog.Info("Created new Datacenter CRD",
		rlog.String("name", name))

	return createdObj, nil
}

// getDatacenterInfoFromConfigMap retrieves datacenter information from the ConfigMap
func getDatacenterInfoFromConfigMap() (name, region, location, zone string) {
	// First try to get from cache/service
	ctx := context.TODO()
	datacenterName, err := datacenternameservice.GetName(ctx)
	if err != nil {
		rlog.Error("Failed to get datacenter name from service", err)
		datacenterName = ""
	}

	// Try to get the ConfigMap directly to extract region, zone and location
	namespace := viper.GetString(consts.NAMESPACE)
	configMapName := viper.GetString(consts.CONFIGMAPNAME)

	if clients.Kubernetes != nil {
		configMap, err := clients.Kubernetes.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
		if err == nil && configMap.Data != nil {
			if datacenterName == "" {
				datacenterName = configMap.Data["name"]
			}
			region = configMap.Data["region"]
			location = configMap.Data["location"]
			zone = configMap.Data["zone"]
		} else {
			rlog.Error("Failed to get ConfigMap for datacenter info", err)
		}
	}

	return datacenterName, region, location, zone
}

// addProviderToDatacenter adds a provider to the specified provider list if it doesn't already exist
func addProviderToDatacenter(datacenterObj *unstructured.Unstructured, providerName, providerType string) {
	datacenterName := datacenterObj.GetName()

	// First, use a read lock to check if the provider already exists
	datacenterRWMutex.RLock()
	// Get the latest version of the datacenter object
	latestObj, err := clients.DynamicClient.Resource(datacenterGVR).Get(context.TODO(), datacenterName, metav1.GetOptions{})
	if err != nil {
		datacenterRWMutex.RUnlock()
		rlog.Error("Failed to get Datacenter CRD", err,
			rlog.String("name", datacenterName))
		return
	}

	// Check if provider already exists in the list
	providers, found, err := unstructured.NestedStringSlice(latestObj.Object, "spec", providerType)
	if err != nil {
		datacenterRWMutex.RUnlock()
		rlog.Error("Failed to get providers from datacenter", err,
			rlog.String("providerType", providerType))
		return
	}

	if !found {
		providers = []string{}
	}

	providerExists := slices.Contains(providers, providerName)
	// If provider already exists, just log and return (no need for a write lock)
	if providerExists {
		datacenterRWMutex.RUnlock()
		rlog.Info("Provider already exists in Datacenter, no update needed",
			rlog.String("providerType", providerType),
			rlog.String("provider", providerName),
			rlog.String("datacenter", datacenterName))
		return
	}

	// Release read lock before acquiring write lock to avoid deadlock
	datacenterRWMutex.RUnlock()

	// Acquire write lock for the update operation
	datacenterRWMutex.Lock()
	defer datacenterRWMutex.Unlock()

	// Get the latest version again after acquiring the write lock
	// This ensures we're working with current data even if it changed while we were waiting for the lock
	latestObj, err = clients.DynamicClient.Resource(datacenterGVR).Get(context.TODO(), datacenterName, metav1.GetOptions{})
	if err != nil {
		rlog.Error("Failed to get updated Datacenter CRD", err,
			rlog.String("name", datacenterName))
		return
	}

	// Re-check if provider exists (in case it was added while we were switching locks)
	providers, found, err = unstructured.NestedStringSlice(latestObj.Object, "spec", providerType)
	if err != nil {
		rlog.Error("Failed to get providers from datacenter", err,
			rlog.String("providerType", providerType))
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
			rlog.Error("Failed to set providers in datacenter", err,
				rlog.String("providerType", providerType))
			return
		}

		// Update the datacenter resource
		_, err = clients.DynamicClient.Resource(datacenterGVR).Update(context.TODO(), latestObj, metav1.UpdateOptions{})
		if err != nil {
			rlog.Error("Failed to update Datacenter CRD", err,
				rlog.String("name", datacenterName))
			return
		}

		rlog.Info("Updated Datacenter CRD with provider",
			rlog.String("name", datacenterName),
			rlog.String("providerType", providerType),
			rlog.String("provider", providerName))
	} else {
		rlog.Info("Provider already exists in Datacenter, no update needed",
			rlog.String("providerType", providerType),
			rlog.String("provider", providerName),
			rlog.String("datacenter", datacenterName))
	}
}

// removeProviderFromDatacenter removes a provider from the specified provider list
func removeProviderFromDatacenter(datacenterObj *unstructured.Unstructured, providerName, providerType string) {
	datacenterName := datacenterObj.GetName()

	// First, use a read lock to check if the provider exists
	datacenterRWMutex.RLock()
	// Get the latest version of the datacenter object
	latestObj, err := clients.DynamicClient.Resource(datacenterGVR).Get(context.TODO(), datacenterName, metav1.GetOptions{})
	if err != nil {
		datacenterRWMutex.RUnlock()
		rlog.Error("Failed to get Datacenter CRD", err,
			rlog.String("name", datacenterName))
		return
	}

	// Check the current providers list
	providers, found, err := unstructured.NestedStringSlice(latestObj.Object, "spec", providerType)
	if err != nil {
		datacenterRWMutex.RUnlock()
		rlog.Error("Failed to get providers from datacenter", err,
			rlog.String("providerType", providerType))
		return
	}

	if !found || len(providers) == 0 {
		// Nothing to remove
		datacenterRWMutex.RUnlock()
		rlog.Info("No providers found in datacenter to remove",
			rlog.String("providerType", providerType),
			rlog.String("datacenter", datacenterName))
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
		datacenterRWMutex.RUnlock()
		rlog.Info("Provider not found in Datacenter, no removal needed",
			rlog.String("providerType", providerType),
			rlog.String("provider", providerName),
			rlog.String("datacenter", datacenterName))
		return
	}

	// Release read lock before acquiring write lock to avoid deadlock
	datacenterRWMutex.RUnlock()

	// Acquire write lock for the update operation
	datacenterRWMutex.Lock()
	defer datacenterRWMutex.Unlock()

	// Get the latest version again after acquiring the write lock
	latestObj, err = clients.DynamicClient.Resource(datacenterGVR).Get(context.TODO(), datacenterName, metav1.GetOptions{})
	if err != nil {
		rlog.Error("Failed to get updated Datacenter CRD", err,
			rlog.String("name", datacenterName))
		return
	}

	// Re-check the providers (in case they changed while we were switching locks)
	providers, found, err = unstructured.NestedStringSlice(latestObj.Object, "spec", providerType)
	if err != nil {
		rlog.Error("Failed to get providers from datacenter", err,
			rlog.String("providerType", providerType))
		return
	}

	if !found || len(providers) == 0 {
		rlog.Info("No providers found in datacenter to remove",
			rlog.String("providerType", providerType),
			rlog.String("datacenter", datacenterName))
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
			rlog.Error("Failed to update providers in datacenter", err,
				rlog.String("providerType", providerType))
			return
		}

		// Update the datacenter resource
		_, err = clients.DynamicClient.Resource(datacenterGVR).Update(context.TODO(), latestObj, metav1.UpdateOptions{})
		if err != nil {
			rlog.Error("Failed to update Datacenter CRD after removal", err,
				rlog.String("name", datacenterName))
			return
		}

		rlog.Info("Removed provider from Datacenter CRD",
			rlog.String("name", datacenterName),
			rlog.String("providerType", providerType),
			rlog.String("provider", providerName))
	} else {
		rlog.Info("Provider not found in Datacenter, no removal needed",
			rlog.String("providerType", providerType),
			rlog.String("provider", providerName),
			rlog.String("datacenter", datacenterName))
	}
}
