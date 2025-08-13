package resourceloglistener

import (
	"github.com/NorskHelsenett/ror/pkg/rlog"
	"github.com/vitistack/datacenter-operator/pkg/eventmanager"
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
	rlog.Info("Resource event occurred",
		rlog.String("type", string(event.Type)),
		rlog.String("kind", event.Resource.GetKind()),
		rlog.String("name", event.Resource.GetName()),
		rlog.String("namespace", event.Resource.GetNamespace()),
		rlog.String("uid", string(event.Resource.GetUID())))
}

// handleKubernetesProviderEvents handles KubernetesProvider-specific events
func handleKubernetesProviderEvents(event eventmanager.ResourceEvent) {
	rlog.Info("KubernetesProvider event occurred",
		rlog.String("type", string(event.Type)),
		rlog.String("name", event.Resource.GetName()),
		rlog.String("uid", string(event.Resource.GetUID())))

	// Here you can add specific logic for handling KubernetesProvider events
	// For example, trigger other systems, update related resources, etc.
}

// handleMachineProviderEvents handles MachineProvider-specific events
func handleMachineProviderEvents(event eventmanager.ResourceEvent) {
	rlog.Info("MachineProvider event occurred",
		rlog.String("type", string(event.Type)),
		rlog.String("name", event.Resource.GetName()),
		rlog.String("uid", string(event.Resource.GetUID())))

	// Here you can add specific logic for handling MachineProvider events
	// For example, trigger other systems, update related resources, etc.
}

func handleConfigMapEvents(event eventmanager.ResourceEvent) {
	rlog.Info("ConfigMap event occurred",
		rlog.String("type", string(event.Type)),
		rlog.String("name", event.Resource.GetName()),
		rlog.String("uid", string(event.Resource.GetUID())))

	// Here you can add specific logic for handling ConfigMap events
	// For example, trigger other systems, update related resources, etc.
}
