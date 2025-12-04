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

	machineName := event.Resource.GetName()
	namespace := event.Resource.GetNamespace()

	vlog.Info("Processing Machine event",
		"type: ", string(event.Type),
		"name: ", machineName,
		"namespace: ", namespace)

	// Get or create the vitistack CRD
	vitistackCrdName := viper.GetString(consts.VITISTACKCRDNAME)
	vitistackObj, err := getOrCreateVitistackCrd(vitistackCrdName)
	if err != nil {
		vlog.Error("Failed to get or create Viti stack CRD", err,
			"name: ", vitistackCrdName)
		return
	}

	// Handle based on event type - we just need to update the count
	switch event.Type {
	case eventmanager.EventAdd:
		incrementMachineCount(vitistackObj)
	case eventmanager.EventDelete:
		decrementMachineCount(vitistackObj)
	case eventmanager.EventUpdate:
		// No count change needed for updates
		vlog.Info("Machine updated, no count change needed",
			"name: ", machineName)
	default:
		vlog.Info("Unhandled event type", "type: ", string(event.Type))
	}
}

// incrementMachineCount increments the activeMachines count in status
func incrementMachineCount(vitistackObj *unstructured.Unstructured) {
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

	// Get current count from status
	currentCount, _, _ := unstructured.NestedInt64(latestObj.Object, "status", "activeMachines")

	// Increment count
	newCount := currentCount + 1

	// Ensure status exists
	status, _, _ := unstructured.NestedMap(latestObj.Object, "status")
	if status == nil {
		status = map[string]any{}
	}

	// Update activeMachines count
	status["activeMachines"] = newCount

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

	vlog.Info("Incremented activeMachines count in Viti stack status",
		"name: ", vitistackName,
		"count: ", newCount)
}

// decrementMachineCount decrements the activeMachines count in status
func decrementMachineCount(vitistackObj *unstructured.Unstructured) {
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

	// Get current count from status
	currentCount, _, _ := unstructured.NestedInt64(latestObj.Object, "status", "activeMachines")

	// Decrement count (don't go below 0)
	newCount := currentCount - 1
	if newCount < 0 {
		newCount = 0
	}

	// Ensure status exists
	status, _, _ := unstructured.NestedMap(latestObj.Object, "status")
	if status == nil {
		status = map[string]any{}
	}

	// Update activeMachines count
	status["activeMachines"] = newCount

	// Set the updated status
	err = unstructured.SetNestedField(latestObj.Object, status, "status")
	if err != nil {
		vlog.Error("Failed to set status in vitistack", err)
		return
	}

	// Update the vitistack resource status
	_, err = k8sclient.DynamicClient.Resource(vitistackGVR).UpdateStatus(context.TODO(), latestObj, metav1.UpdateOptions{})
	if err != nil {
		vlog.Error("Failed to update Viti stack CRD status after decrement", err,
			"name: ", vitistackName)
		return
	}

	vlog.Info("Decremented activeMachines count in Viti stack status",
		"name: ", vitistackName,
		"count: ", newCount)
}
