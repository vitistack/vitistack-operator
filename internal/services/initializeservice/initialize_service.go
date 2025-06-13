package initializeservice

import (
	"context"
	"fmt"
	"strings"

	"github.com/NorskHelsenett/ror/pkg/rlog"
	"github.com/spf13/viper"
	"github.com/vitistack/datacenter-operator/internal/clients"
	"github.com/vitistack/datacenter-operator/pkg/consts"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CheckPrerequisites() {
	rlog.Info("Running prerequisite checks...")
	CheckConfigmap()
	CheckCRDs()
	rlog.Info("✅ Prerequisite checks passed")
}

// CheckConfigmap verifies that the required ConfigMap exists and has the necessary configuration
func CheckConfigmap() {
	var errors []string
	kubernetesclient := clients.Kubernetes

	// Define the namespace and name of the ConfigMap we expect to exist
	namespace := viper.GetString(consts.NAMESPACE)         // This might need to be adjusted based on your deployment
	configMapName := viper.GetString(consts.CONFIGMAPNAME) // Change this to your specific ConfigMap name

	// Attempt to get the ConfigMap
	configMap, err := kubernetesclient.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to find required ConfigMap '%s' in namespace '%s': %s", configMapName, namespace, err.Error()))
	} else {
		rlog.Info(fmt.Sprintf("✅ ConfigMap '%s' found in namespace '%s'", configMapName, namespace))

		// Verify that the ConfigMap contains required keys
		// Add your specific key checks here based on what your application needs
		requiredKeys := []string{"name", "provider", "region", "location"} // Example required keys
		for _, key := range requiredKeys {
			if _, exists := configMap.Data[key]; !exists {
				errors = append(errors, fmt.Sprintf("ConfigMap '%s' is missing required key: %s", configMapName, key))
			} else {
				rlog.Info(fmt.Sprintf("✅ ConfigMap contains required key: %s", key))
			}
		}
	}

	// If we collected any errors, report them all together
	if len(errors) > 0 {
		errorMessage := fmt.Sprintf("ConfigMap prerequisite checks failed:\n- %s", strings.Join(errors, "\n- "))
		rlog.Error(errorMessage, nil)
		panic(errorMessage)
	}

	rlog.Info("✅ ConfigMap prerequisite checks passed")
}

func CheckCRDs() {
	var errors []string
	kubernetesclient := clients.Kubernetes

	// Check if the cluster is accessible by listing namespaces
	namespaces, err := kubernetesclient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to access cluster: %s", err.Error()))
	} else if len(namespaces.Items) == 0 {
		errors = append(errors, "No namespaces found in the cluster")
	}

	// Only continue with CRD checks if we can access the cluster
	if len(errors) == 0 {
		// Check if all required CRDs are installed
		resources, err := kubernetesclient.Discovery().ServerResourcesForGroupVersion("vitistack.io/v1alpha1")
		if err != nil {
			errors = append(errors, fmt.Sprintf("CRDs are not installed properly: %s", err.Error()))
		} else if len(resources.APIResources) == 0 {
			errors = append(errors, "No resources found for the required CRDs")
		} else {
			// Check for each required CRD
			requiredCRDs := []string{"KubernetesProvider", "MachineProvider", "Datacenter"}
			for _, crdKind := range requiredCRDs {
				if !crdExists(resources, crdKind) {
					errors = append(errors, fmt.Sprintf("%s CRD is not installed", crdKind))
				} else {
					rlog.Info(fmt.Sprintf("✅ %s CRD is installed", crdKind))
				}
			}
		}
	}

	// If we collected any errors, report them all together
	if len(errors) > 0 {
		errorMessage := fmt.Sprintf("Prerequisite checks failed:\n- %s", strings.Join(errors, "\n- "))
		rlog.Error(errorMessage, nil)
		panic(errorMessage)
	}

	rlog.Info("✅ All prerequisite checks passed")
}

// crdExists verifies that a specific CRD exists in the API resources
func crdExists(resources *metav1.APIResourceList, crdKind string) bool {
	for _, resource := range resources.APIResources {
		if resource.Kind == crdKind {
			return true
		}
	}
	return false
}
