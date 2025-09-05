package resourceloglistener

import (
	"github.com/vitistack/common/pkg/loggers/vlog"
	"github.com/vitistack/vitistack-operator/pkg/eventmanager"
)

// RegisterListeners registers all resource event listeners with the event bus
func RegisterListeners() {
	// Register a global listener for all resource events
	eventmanager.EventBus.SubscribeAll(logAllEvents)

	// Register specific listeners for different resource kinds
	eventmanager.EventBus.Subscribe("KubernetesProvider", handleKubernetesProviderEvents)
	eventmanager.EventBus.Subscribe("MachineProvider", handleMachineProviderEvents)
	eventmanager.EventBus.Subscribe("ConfigMap", handleConfigMapEvents)
}

// logAllEvents logs all resource events (example of a global event handler)
func logAllEvents(event eventmanager.ResourceEvent) {
	vlog.Info("Resource event occurred",
		"type", string(event.Type),
		"kind", event.Resource.GetKind(),
		"name", event.Resource.GetName(),
		"namespace", event.Resource.GetNamespace(),
		"uid", string(event.Resource.GetUID()))
}

// handleKubernetesProviderEvents handles KubernetesProvider-specific events
func handleKubernetesProviderEvents(event eventmanager.ResourceEvent) {
	vlog.Info("KubernetesProvider event occurred",
		"type", string(event.Type),
		"name", event.Resource.GetName(),
		"uid", string(event.Resource.GetUID()))

	// Here you can add specific logic for handling KubernetesProvider events
	// For example, trigger other systems, update related resources, etc.
}

// handleMachineProviderEvents handles MachineProvider-specific events
func handleMachineProviderEvents(event eventmanager.ResourceEvent) {
	vlog.Info("MachineProvider event occurred",
		"type", string(event.Type),
		"name", event.Resource.GetName(),
		"uid", string(event.Resource.GetUID()))

	// Here you can add specific logic for handling MachineProvider events
	// For example, trigger other systems, update related resources, etc.
}

func handleConfigMapEvents(event eventmanager.ResourceEvent) {
	vlog.Info("ConfigMap event occurred",
		"type", string(event.Type),
		"name", event.Resource.GetName(),
		"uid", string(event.Resource.GetUID()))

	// Here you can add specific logic for handling ConfigMap events
	// For example, trigger other systems, update related resources, etc.
}
