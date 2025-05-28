package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineProvider is the Schema for the MachineProviders API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=machineproviders,scope=Namespaced,shortName=mp
type MachineProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineProviderSpec   `json:"spec,omitempty"`
	Status MachineProviderStatus `json:"status,omitempty"`
}

// MachineProviderSpec defines the desired state of MachineProvider
type MachineProviderSpec struct {
	Config MachineProviderConfig `json:"config"`
}

type MachineProviderConfig struct {
	Name string `json:"name"`
}

// MachineProviderStatus defines the observed state of MachineProvider
type MachineProviderStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// MachineProviderList contains a list of MachineProvider
type MachineProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MachineProvider{}, &MachineProviderList{})
}
