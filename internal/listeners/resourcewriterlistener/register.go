package resourcewriterlistener

import (
	"github.com/vitistack/vitistack-operator/pkg/eventmanager"
)

// RegisterWriters registers all resource event writers with the event bus
func RegisterWriters() {
	eventmanager.EventBus.Subscribe("KubernetesProvider", handleKubernetesProviderEvents)
	eventmanager.EventBus.Subscribe("MachineProvider", handleMachineProviderEvents)
	eventmanager.EventBus.Subscribe("MachineClass", handleMachineClassEvents)
	eventmanager.EventBus.Subscribe("KubernetesCluster", handleKubernetesClusterEvents)
	eventmanager.EventBus.Subscribe("Machine", handleMachineEvents)
	eventmanager.EventBus.Subscribe("ConfigMap", handleConfigMapEvents)
}
