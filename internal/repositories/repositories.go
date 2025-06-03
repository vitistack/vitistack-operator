package repositories

import (
	"github.com/vitistack/crds/pkg/v1alpha1"
	"github.com/vitistack/datacenter-operator/internal/repositories/kubernetesproviderrepository"
	"github.com/vitistack/datacenter-operator/internal/repositories/machineproviderrepository"
	"github.com/vitistack/datacenter-operator/internal/repositoryinterfaces"
)

var (
	KubernetesProviderRepository repositoryinterfaces.Repository[v1alpha1.KubernetesProvider]
	MachineProviderRepository    repositoryinterfaces.Repository[v1alpha1.MachineProvider]
)

func InitializeRepositories() {
	MachineProviderRepository = machineproviderrepository.NewMachineProviderRepository()
	KubernetesProviderRepository = kubernetesproviderrepository.NewKubernetesProviderRepository()
}
