package dynamichandler

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (handler) GetSchemas() []schema.GroupVersionResource {
	return []schema.GroupVersionResource{
		{
			Group:    "vitistack.io",
			Version:  "v1alpha1",
			Resource: "kubernetesproviders",
		},
		{
			Group:    "vitistack.io",
			Version:  "v1alpha1",
			Resource: "machineproviders",
		},
		{
			Group:    "vitistack.io",
			Version:  "v1alpha1",
			Resource: "machineclasses",
		},
		{
			Group:    "vitistack.io",
			Version:  "v1alpha1",
			Resource: "kubernetesclusters",
		},
		{
			Group:    "vitistack.io",
			Version:  "v1alpha1",
			Resource: "machines",
		},
		{
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
		},
	}
}
