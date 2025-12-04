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
	eventmanager.EventBus.Subscribe("MachineClass", handleMachineClassEvents)
	eventmanager.EventBus.Subscribe("KubernetesCluster", handleKubernetesClusterEvents)
	eventmanager.EventBus.Subscribe("Machine", handleMachineEvents)
	eventmanager.EventBus.Subscribe("ConfigMap", handleConfigMapEvents)
}

// logAllEvents logs all resource events (example of a global event handler)
func logAllEvents(event eventmanager.ResourceEvent) {
	if event.Resource == nil {
		vlog.Error("Resource is nil in event", nil)
		return
	}
	vlog.Info("Resource event occurred",
		"type", string(event.Type),
		"kind", event.Resource.GetKind(),
		"name", event.Resource.GetName(),
		"namespace", event.Resource.GetNamespace(),
		"uid", string(event.Resource.GetUID()))
}

// handleKubernetesProviderEvents handles KubernetesProvider-specific events
func handleKubernetesProviderEvents(event eventmanager.ResourceEvent) {
	if event.Resource == nil {
		return
	}
	vlog.Info("KubernetesProvider event occurred",
		"type", string(event.Type),
		"name", event.Resource.GetName(),
		"uid", string(event.Resource.GetUID()))
}

// handleMachineProviderEvents handles MachineProvider-specific events
func handleMachineProviderEvents(event eventmanager.ResourceEvent) {
	if event.Resource == nil {
		return
	}
	vlog.Info("MachineProvider event occurred",
		"type", string(event.Type),
		"name", event.Resource.GetName(),
		"uid", string(event.Resource.GetUID()))
}

func handleConfigMapEvents(event eventmanager.ResourceEvent) {
	if event.Resource == nil {
		return
	}
	vlog.Info("ConfigMap event occurred",
		"type", string(event.Type),
		"name", event.Resource.GetName(),
		"uid", string(event.Resource.GetUID()))
}

// handleMachineClassEvents handles MachineClass-specific events
func handleMachineClassEvents(event eventmanager.ResourceEvent) {
	if event.Resource == nil {
		return
	}
	vlog.Info("MachineClass event occurred",
		"type", string(event.Type),
		"name", event.Resource.GetName(),
		"uid", string(event.Resource.GetUID()))
}

// handleKubernetesClusterEvents handles KubernetesCluster-specific events
func handleKubernetesClusterEvents(event eventmanager.ResourceEvent) {
	if event.Resource == nil {
		return
	}
	vlog.Info("KubernetesCluster event occurred",
		"type", string(event.Type),
		"name", event.Resource.GetName(),
		"namespace", event.Resource.GetNamespace(),
		"uid", string(event.Resource.GetUID()))
}

// handleMachineEvents handles Machine-specific events
func handleMachineEvents(event eventmanager.ResourceEvent) {
	if event.Resource == nil {
		return
	}
	vlog.Info("Machine event occurred",
		"type", string(event.Type),
		"name", event.Resource.GetName(),
		"namespace", event.Resource.GetNamespace(),
		"uid", string(event.Resource.GetUID()))
}
