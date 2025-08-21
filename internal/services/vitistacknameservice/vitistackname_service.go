package vitistacknameservice

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/NorskHelsenett/ror/pkg/rlog"
	"github.com/spf13/viper"
	"github.com/vitistack/vitistack-operator/internal/cache"
	"github.com/vitistack/vitistack-operator/internal/clients"
	"github.com/vitistack/vitistack-operator/pkg/consts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Required fields in the configmap that must be present
var requiredConfigMapFields = []string{"name", "location", "zone"}

func GetName(ctx context.Context) (string, error) {
	configMapName := viper.GetString(consts.CONFIGMAPNAME)
	namespace := viper.GetString(consts.NAMESPACE)

	rlog.Info("Getting datacenter name",
		rlog.String("configMapName", configMapName),
		rlog.String("namespace", namespace))

	// Try to get from cache first
	configData, err := getConfigDataFromCache(ctx, namespace, configMapName)
	if err == nil {
		rlog.Info("Retrieved ConfigMap data from cache",
			rlog.String("name", configData["name"]))
		// If cache is valid, return the data
		return configData["name"], nil
	}

	rlog.Info("Cache miss or error, falling back to Kubernetes API", rlog.Any("error", err))

	// If cache fails, get from Kubernetes API and update cache
	configData, err = getConfigDataFromK8s(ctx, namespace, configMapName)
	if err != nil {
		return "", fmt.Errorf("failed to get config data: %w", err)
	}

	rlog.Info("Retrieved ConfigMap data from Kubernetes API",
		rlog.String("name", configData["name"]))

	// Update cache with fresh data from Kubernetes API
	cacheKey := buildCacheKey(namespace, configMapName)
	configMap, err := getConfigMapFromK8s(ctx, namespace, configMapName)
	if err == nil && configMap != nil {
		err = cache.Cache.Set(ctx, cacheKey, configMap)
		if err != nil {
			rlog.Error("Failed to update cache with fresh ConfigMap data:", err)
		} else {
			rlog.Info("Updated cache with fresh ConfigMap data")
		}
	}

	return configData["name"], nil
}

// InvalidateCache removes the ConfigMap from cache to force fresh data retrieval
func InvalidateCache(ctx context.Context, namespace, name string) error {
	cacheKey := buildCacheKey(namespace, name)
	err := cache.Cache.Delete(ctx, cacheKey)
	if err != nil {
		rlog.Error("Failed to invalidate cache:", err)
		return err
	}
	rlog.Info("Cache invalidated successfully", rlog.String("key", cacheKey))
	return nil
}

// buildCacheKey creates a standardized cache key for config maps
func buildCacheKey(namespace, name string) string {
	return fmt.Sprintf("configmap-%s-%s", namespace, name)
}

// extractConfigDataFromCache extracts config map data from the cached JSON string
func extractConfigDataFromCache(cachedDataStr string) (map[string]string, error) {
	if cachedDataStr == "" {
		return nil, fmt.Errorf("empty cache data")
	}

	// Parse the JSON string into a map
	var configMapData map[string]interface{}
	err := json.Unmarshal([]byte(cachedDataStr), &configMapData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	// Extract the data field which contains our configmap values
	data, ok := configMapData["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid cache data structure")
	}

	// Convert string interface values to actual strings
	configData := make(map[string]string)
	for k, v := range data {
		if strValue, ok := v.(string); ok {
			configData[k] = strValue
		}
	}

	return configData, nil
}

// validateConfigData checks that all required fields are present in the data
func validateConfigData(data map[string]string) error {
	for _, field := range requiredConfigMapFields {
		if data[field] == "" {
			return fmt.Errorf("no %s found in data", field)
		}
	}
	return nil
}

// getConfigMapFromK8s fetches a ConfigMap from Kubernetes
func getConfigMapFromK8s(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error) {
	return clients.Kubernetes.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
}

// getConfigDataFromCache attempts to retrieve config data from the cache
func getConfigDataFromCache(ctx context.Context, namespace, name string) (map[string]string, error) {
	cacheKey := buildCacheKey(namespace, name)
	cachedDataStr, err := cache.Cache.Get(ctx, cacheKey)
	if err != nil {
		rlog.Error("Cache miss:", err)
		return nil, err
	}

	// If cachedData is empty, return an error
	if cachedDataStr == "" {
		return nil, fmt.Errorf("no data in cache")
	}

	configData, err := extractConfigDataFromCache(cachedDataStr)
	if err != nil {
		rlog.Error("Failed to process cached data:", err)
		return nil, err
	}

	err = validateConfigData(configData)
	if err != nil {
		rlog.Error("Invalid cached data:", err)
		return nil, err
	}

	return configData, nil
}

// getConfigDataFromK8s gets config data directly from Kubernetes
func getConfigDataFromK8s(ctx context.Context, namespace, name string) (map[string]string, error) {
	fmt.Println("Retrieving configmap from Kubernetes API")

	configMap, err := getConfigMapFromK8s(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve configmap: %w", err)
	}

	if configMap == nil {
		return nil, fmt.Errorf("configmap not found")
	}

	if configMap.Data == nil {
		return nil, fmt.Errorf("no data found in configmap")
	}

	err = validateConfigData(configMap.Data)
	if err != nil {
		return nil, err
	}

	return configMap.Data, nil
}
