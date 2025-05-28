package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Datacenter is the Schema for the Datacenters API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=datacenters,scope=Namespaced,shortName=dc
type Datacenter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatacenterSpec   `json:"spec,omitempty"`
	Status DatacenterStatus `json:"status,omitempty"`
}

// DatacenterSpec defines the desired state of Datacenter
type DatacenterSpec struct {
	Name                string   `json:"name"`
	MachineProviders    []string `json:"machineProviders"`
	KubernetesProviders []string `json:"kubernetesProviders"`
}

// DatacenterStatus defines the observed state of Datacenter
type DatacenterStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DatacenterList contains a list of Datacenter
type DatacenterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Datacenter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Datacenter{}, &DatacenterList{})
}
