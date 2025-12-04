package resourcewriterlistener

import (
	"context"

	"github.com/spf13/viper"
	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/vitistack-operator/pkg/consts"
	"github.com/vitistack/vitistack-operator/pkg/eventmanager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// handleMachineEvents processes events for Machine resources
func handleMachineEvents(event eventmanager.ResourceEvent) {
	if event.Resource == nil {
		vlog.Error("Resource is nil in Machine event", nil)
		return
	}
	updateVitistackStatusWithMachine(event)
}

// updateVitistackStatusWithMachine handles updating the Vitistack CRD status with machine count
func updateVitistackStatusWithMachine(event eventmanager.ResourceEvent) {
	// Use the shared dynamic client
	if k8sclient.DynamicClient == nil {
		vlog.Error("Dynamic client is not initialized", nil)
		return
	}

	// Only update count on Add or Delete events
	if event.Type != eventmanager.EventAdd && event.Type != eventmanager.EventDelete {
		return
	}

	// Get or create the vitistack CRD
	vitistackCrdName := viper.GetString(consts.VITISTACKCRDNAME)
	vitistackObj, err := getOrCreateVitistackCrd(vitistackCrdName)
	if err != nil {
		vlog.Error("Failed to get or create Viti stack CRD", err,
			"name: ", vitistackCrdName)
		return
	}

	// Count actual machines from cluster and update status
	updateMachineCount(vitistackObj)
}

// updateMachineCount counts actual machines from the cluster and updates the status
func updateMachineCount(vitistackObj *unstructured.Unstructured) {
	vitistackName := vitistackObj.GetName()

	// List all machines from the cluster
	machineList, err := k8sclient.DynamicClient.Resource(machineGVR).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		vlog.Error("Failed to list machines", err)
		return
	}

	// Count the machines
	actualCount := int64(len(machineList.Items))

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

	// Get current count from status
	currentCount, _, _ := unstructured.NestedInt64(latestObj.Object, "status", "activeMachines")

	// Only update if count has changed
	if currentCount == actualCount {
		return
	}

	// Ensure status exists
	status, _, _ := unstructured.NestedMap(latestObj.Object, "status")
	if status == nil {
		status = map[string]any{}
	}

	// Update activeMachines count
	status["activeMachines"] = actualCount

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

	vlog.Info("Updated activeMachines count in Viti stack status",
		"name: ", vitistackName,
		"previousCount: ", currentCount,
		"newCount: ", actualCount)
}
