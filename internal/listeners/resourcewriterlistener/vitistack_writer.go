package resourcewriterlistener

import (
	"context"
	"time"

	"github.com/spf13/viper"
	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/vitistack-operator/internal/services/vitistacknameservice"
	"github.com/vitistack/vitistack-operator/pkg/consts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// getOrCreateVitistackCrd tries to get an existing vitistack CRD or creates a new one if it doesn't exist
func getOrCreateVitistackCrd(name string) (*unstructured.Unstructured, error) {
	// Use a read lock first since we're just checking if the vitistack exists
	vitistackRWMutex.RLock()
	vitistackObj, err := k8sclient.DynamicClient.Resource(vitistackGVR).Get(context.TODO(), name, metav1.GetOptions{})
	vitistackRWMutex.RUnlock()

	if err == nil {
		return vitistackObj, nil
	}

	// Get vitistack information from ConfigMap
	vitistackName, region, country, zone, infrastructure := getVitistackInfoFromConfigMap()

	// If the vitistack doesn't exist, we need to create it - acquire a write lock
	vitistackRWMutex.Lock()
	defer vitistackRWMutex.Unlock()

	// Check again in case another goroutine created the vitistack while we were waiting for the lock
	vitistackObj, err = k8sclient.DynamicClient.Resource(vitistackGVR).Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil {
		return vitistackObj, nil
	}

	// Build location object if location is provided
	var locationObj map[string]any
	if country != "" {
		locationObj = map[string]any{
			"country": country,
		}
	}

	spec := map[string]any{
		"displayName":    vitistackName,
		"infrastructure": infrastructure,
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

	// Initialize empty status with empty lists for providers, machineClasses, and clusters
	status := map[string]any{
		"kubernetesProviders":     []interface{}{},
		"machineProviders":        []interface{}{},
		"machineClasses":          []interface{}{},
		"clusters":                []interface{}{},
		"providerStatuses":        []interface{}{},
		"conditions":              []interface{}{},
		"kubernetesProviderCount": int64(0),
		"machineProviderCount":    int64(0),
		"activeClusters":          int64(0),
		"activeMachines":          int64(0),
		"phase":                   "Initializing",
		"lastReconcileTime":       time.Now().UTC().Format(time.RFC3339),
		"observedGeneration":      int64(1),
	}

	// Add ConfigMap-derived fields to status if available
	if vitistackName != "" {
		status["displayName"] = vitistackName
	}
	if region != "" {
		status["region"] = region
	}
	if zone != "" {
		status["zone"] = zone
	}
	if infrastructure != "" {
		status["infrastructure"] = infrastructure
	}
	if locationObj != nil {
		status["location"] = locationObj
	}

	err = unstructured.SetNestedField(createdObj.Object, status, "status")
	if err != nil {
		vlog.Error("Failed to set initial status in vitistack", err)
		return createdObj, nil // Return the created object even if status update fails
	}

	// Update status
	updatedObj, err := k8sclient.DynamicClient.Resource(vitistackGVR).UpdateStatus(context.TODO(), createdObj, metav1.UpdateOptions{})
	if err != nil {
		vlog.Error("Failed to update initial Vitistack status", err)
		return createdObj, nil // Return the created object even if status update fails
	}

	vlog.Info("Initialized Vitistack CRD status with empty provider lists",
		"name: ", name)

	return updatedObj, nil
}

// getVitistackInfoFromConfigMap retrieves vitistack information from the ConfigMap
func getVitistackInfoFromConfigMap() (name, region, country, zone, infrastructure string) {
	// First try to get from cache/service
	ctx := context.TODO()
	vitistackName, err := vitistacknameservice.GetName(ctx)
	if err != nil {
		vlog.Error("Failed to get vitistack name from service: ", err)
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
			infrastructure = configMap.Data["infrastructure"]
		} else {
			vlog.Error("Failed to get ConfigMap for vitistack info", err)
		}
	}

	return vitistackName, region, country, zone, infrastructure
}
