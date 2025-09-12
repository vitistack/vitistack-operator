package initializeservice

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/common/pkg/operator/crdcheck"
	"github.com/vitistack/vitistack-operator/pkg/consts"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CheckPrerequisites() {
	vlog.Info("Running prerequisite checks...")
	CheckConfigmap()
	CheckCRDs()
	vlog.Info("✅ Prerequisite checks passed")
}

// CheckConfigmap verifies that the required ConfigMap exists and has the necessary configuration
func CheckConfigmap() {
	var errors []string
	kubernetesclient := k8sclient.Kubernetes

	// Define the namespace and name of the ConfigMap we expect to exist
	namespace := viper.GetString(consts.NAMESPACE)         // This might need to be adjusted based on your deployment
	configMapName := viper.GetString(consts.CONFIGMAPNAME) // Change this to your specific ConfigMap name

	// Attempt to get the ConfigMap
	configMap, err := kubernetesclient.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to find required ConfigMap '%s' in namespace '%s': %s", configMapName, namespace, err.Error()))
	} else {
		vlog.Info(fmt.Sprintf("✅ ConfigMap '%s' found in namespace '%s'", configMapName, namespace))

		// Verify that the ConfigMap contains required keys
		// Add your specific key checks here based on what your application needs
		requiredKeys := []string{"name", "provider", "region", "location"} // Example required keys
		for _, key := range requiredKeys {
			if _, exists := configMap.Data[key]; !exists {
				errors = append(errors, fmt.Sprintf("ConfigMap '%s' is missing required key: %s", configMapName, key))
			} else {
				vlog.Info(fmt.Sprintf("✅ ConfigMap contains required key: %s", key))
			}
		}
	}

	// If we collected any errors, report them all together
	if len(errors) > 0 {
		errorMessage := fmt.Sprintf("ConfigMap prerequisite checks failed:\n- %s", strings.Join(errors, "\n- "))
		vlog.Error(errorMessage, nil)
		panic(errorMessage)
	}

	vlog.Info("✅ ConfigMap prerequisite checks passed")
}

func CheckCRDs() {
	vlog.Info("Checking required CRDs...")

	crdcheck.MustEnsureInstalled(context.TODO(),
		// your CRD plural
		crdcheck.Ref{Group: "vitistack.io", Version: "v1alpha1", Resource: "machines"},
		crdcheck.Ref{Group: "vitistack.io", Version: "v1alpha1", Resource: "kubernetesclusters"},
		crdcheck.Ref{Group: "vitistack.io", Version: "v1alpha1", Resource: "kubernetesproviders"},
		crdcheck.Ref{Group: "vitistack.io", Version: "v1alpha1", Resource: "machineproviders"},
		crdcheck.Ref{Group: "vitistack.io", Version: "v1alpha1", Resource: "networknamespaces"},
		crdcheck.Ref{Group: "vitistack.io", Version: "v1alpha1", Resource: "networkconfigurations"},
		crdcheck.Ref{Group: "vitistack.io", Version: "v1alpha1", Resource: "vitistack"},
	)

	vlog.Info("✅ All crds checks passed")
}
