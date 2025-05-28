package repositories

import (
	"github.com/NorskHelsenett/oss-datacenter-operator/internal/repositories/kubernetesproviderrepository"
	"github.com/NorskHelsenett/oss-datacenter-operator/internal/repositories/machineproviderrepository"
	"github.com/NorskHelsenett/oss-datacenter-operator/internal/repositoryinterfaces"
	"github.com/NorskHelsenett/oss-datacenter-operator/pkg/crds/v1alpha1"
)

var (
	KubernetesProviderRepository repositoryinterfaces.Repository[v1alpha1.KubernetesProvider]
	MachineProviderRepository    repositoryinterfaces.Repository[v1alpha1.MachineProvider]
)

func InitializeRepositories() {
	MachineProviderRepository = machineproviderrepository.NewMachineProviderRepository()
	KubernetesProviderRepository = kubernetesproviderrepository.NewKubernetesProviderRepository()
}
