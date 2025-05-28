package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KubernetesProvider is the Schema for the KubernetesProviders API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=kubernetesproviders,scope=Namespaced,shortName=kp
type KubernetesProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubernetesProviderSpec   `json:"spec,omitempty"`
	Status KubernetesProviderStatus `json:"status,omitempty"`
}

// KubernetesProviderSpec defines the desired state of KuberneteProvider
type KubernetesProviderSpec struct {
	Config KubernetesProviderConfig `json:"config"`
}

type KubernetesProviderConfig struct {
	// The name of the Kubernetes provider
	Name string `json:"name"`
}

// KubernetesProviderStatus defines the observed state of KuberneteProvider
type KubernetesProviderStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// KubernetesProviderList contains a list of KuberneteProvider
type KubernetesProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubernetesProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubernetesProvider{}, &KubernetesProviderList{})
}
