package resourcewriterlistener

import (
	"context"
	"slices"

	"github.com/spf13/viper"
	"github.com/vitistack/common/pkg/clients/k8sclient"
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/vitistack-operator/pkg/consts"
	"github.com/vitistack/vitistack-operator/pkg/eventmanager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// handleMachineClassEvents processes events for MachineClass resources
func handleMachineClassEvents(event eventmanager.ResourceEvent) {
	if event.Resource == nil {
		vlog.Error("Resource is nil in MachineClass event", nil)
		return
	}
	updateVitistackStatusWithMachineClass(event)
}

// updateVitistackStatusWithMachineClass handles updating the Viti stack CRD status with machine class information
func updateVitistackStatusWithMachineClass(event eventmanager.ResourceEvent) {
	// Use the shared dynamic client
	if k8sclient.DynamicClient == nil {
		vlog.Error("Dynamic client is not initialized", nil)
		return
	}

	machineClassName := event.Resource.GetName()

	vlog.Info("Processing MachineClass event",
		"type: ", string(event.Type),
		"name: ", machineClassName)

	// Get or create the vitistack CRD
	vitistackCrdName := viper.GetString(consts.VITISTACKCRDNAME)
	vitistackObj, err := getOrCreateVitistackCrd(vitistackCrdName)
	if err != nil {
		vlog.Error("Failed to get or create Viti stack CRD", err,
			"name: ", vitistackCrdName)
		return
	}

	// Handle based on event type
	switch event.Type {
	case eventmanager.EventAdd, eventmanager.EventUpdate:
		addMachineClassToVitistackStatus(vitistackObj, machineClassName)
	case eventmanager.EventDelete:
		removeMachineClassFromVitistackStatus(vitistackObj, machineClassName)
	default:
		vlog.Info("Unhandled event type", "type: ", string(event.Type))
	}
}

// addMachineClassToVitistackStatus adds a machine class to the status if it doesn't already exist
func addMachineClassToVitistackStatus(vitistackObj *unstructured.Unstructured, machineClassName string) {
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

	// Get current machine classes from status (as array of strings)
	machineClasses, found, err := unstructured.NestedStringSlice(latestObj.Object, "status", "machineClasses")
	if err != nil {
		vlog.Error("Failed to get machineClasses from vitistack status", err)
		return
	}

	if !found {
		machineClasses = []string{}
	}

	// Check if machine class already exists
	for _, mc := range machineClasses {
		if mc == machineClassName {
			vlog.Info("MachineClass already exists in Viti stack status, no update needed",
				"machineClass: ", machineClassName,
				"vitistack: ", vitistackName)
			return
		}
	}

	// Add the machine class
	machineClasses = append(machineClasses, machineClassName)

	// Convert []string to []interface{} for SetNestedField compatibility
	machineClassesInterface := make([]interface{}, len(machineClasses))
	for i, mc := range machineClasses {
		machineClassesInterface[i] = mc
	}

	// Ensure status exists
	status, _, _ := unstructured.NestedMap(latestObj.Object, "status")
	if status == nil {
		status = map[string]any{}
	}

	// Update machineClasses in status
	status["machineClasses"] = machineClassesInterface

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

	vlog.Info("Added MachineClass to Viti stack status",
		"name: ", vitistackName,
		"machineClass: ", machineClassName)
}

// removeMachineClassFromVitistackStatus removes a machine class from the status
func removeMachineClassFromVitistackStatus(vitistackObj *unstructured.Unstructured, machineClassName string) {
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

	// Get current machine classes from status
	machineClasses, found, err := unstructured.NestedStringSlice(latestObj.Object, "status", "machineClasses")
	if err != nil {
		vlog.Error("Failed to get machineClasses from vitistack status", err)
		return
	}

	if !found || len(machineClasses) == 0 {
		vlog.Info("No machineClasses found in vitistack status to remove",
			"vitistack: ", vitistackName)
		return
	}

	// Find and remove the machine class
	machineClassIndex := -1
	for i, mc := range machineClasses {
		if mc == machineClassName {
			machineClassIndex = i
			break
		}
	}

	if machineClassIndex < 0 {
		vlog.Info("MachineClass not found in Viti stack status, no removal needed",
			"machineClass: ", machineClassName,
			"vitistack: ", vitistackName)
		return
	}

	// Remove the machine class
	machineClasses = slices.Delete(machineClasses, machineClassIndex, machineClassIndex+1)

	// Convert []string to []interface{} for SetNestedField compatibility
	machineClassesInterface := make([]interface{}, len(machineClasses))
	for i, mc := range machineClasses {
		machineClassesInterface[i] = mc
	}

	// Ensure status exists
	status, _, _ := unstructured.NestedMap(latestObj.Object, "status")
	if status == nil {
		status = map[string]any{}
	}

	// Update machineClasses in status
	status["machineClasses"] = machineClassesInterface

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

	vlog.Info("Removed MachineClass from Viti stack status",
		"name: ", vitistackName,
		"machineClass: ", machineClassName)
}
