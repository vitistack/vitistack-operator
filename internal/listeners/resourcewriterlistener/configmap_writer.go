package resourcewriterlistener

import (
	"context"

	"github.com/spf13/viper"
	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/vitistack-operator/internal/services/vitistacknameservice"
	"github.com/vitistack/vitistack-operator/pkg/consts"
	"github.com/vitistack/vitistack-operator/pkg/eventmanager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// handleConfigMapEvents processes events for ConfigMap resources
func handleConfigMapEvents(event eventmanager.ResourceEvent) {
	if event.Resource == nil {
		vlog.Error("Resource is nil in ConfigMap event", nil)
		return
	}
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

	updateVitistackFromConfigMap(event)
}

// updateVitistackFromConfigMap updates the Vitistack CRD based on ConfigMap data
func updateVitistackFromConfigMap(event eventmanager.ResourceEvent) {
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
	vitistackObj, err := getOrCreateVitistackCrd(vitistackCrdName)
	if err != nil {
		vlog.Error("Failed to get or create Viti stack CRD", err,
			"name", event.Resource.GetName(),
			"namespace", event.Resource.GetNamespace())
		return
	}

	// Acquire write lock for the update operation
	vitistackRWMutex.Lock()
	defer vitistackRWMutex.Unlock()

	// Get the latest version of the vitistack
	latestObj, err := k8sclient.DynamicClient.Resource(vitistackGVR).Get(context.TODO(), vitistackObj.GetName(), metav1.GetOptions{})
	if err != nil {
		vlog.Error("Failed to get latest Viti stack CRD", err)
		return
	}

	// Get current status or create new one
	status, _, _ := unstructured.NestedMap(latestObj.Object, "status")
	if status == nil {
		status = map[string]any{}
	}

	// Initialize missing status fields with defaults
	initializeStatusDefaults(status)

	// Track if any updates are needed
	updateNeeded := false

	// Update displayName in status
	if vitistackName != "" {
		status["displayName"] = vitistackName
		updateNeeded = true
	}

	// Update region in status
	if region != "" {
		status["region"] = region
		updateNeeded = true
	}

	// Update zone in status
	if zone != "" {
		status["zone"] = zone
		updateNeeded = true
	}

	// Update location in status (as a nested object with country field)
	if country != "" {
		locationObj := map[string]any{
			"country": country,
		}
		status["location"] = locationObj
		updateNeeded = true
	}

	// Update description in status
	if description != "" {
		status["description"] = description
		updateNeeded = true
	}

	// Update infrastructure in status
	if infrastructure != "" {
		status["infrastructure"] = infrastructure
		updateNeeded = true
	}

	// Only update if changes were made
	if !updateNeeded {
		vlog.Info("No updates needed for Viti stack CRD")
		return
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
			"name", event.Resource.GetName(),
			"namespace", event.Resource.GetNamespace())
		return
	}

	vlog.Info("Updated Viti stack CRD status from ConfigMap",
		"vitistackName: ", vitistackName,
		"region: ", region,
		"zone: ", zone,
		"country: ", country,
		"provider: ", provider,
		"namespace: ", event.Resource.GetNamespace())
}
