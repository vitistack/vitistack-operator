package repositories

import (
	"github.com/vitistack/common/pkg/v1alpha1"
	"github.com/vitistack/vitistack-operator/internal/repositories/kubernetesproviderrepository"
	"github.com/vitistack/vitistack-operator/internal/repositories/machineproviderrepository"
	"github.com/vitistack/vitistack-operator/internal/repositoryinterfaces"
)

var (
	KubernetesProviderRepository repositoryinterfaces.Repository[v1alpha1.KubernetesProvider]
	MachineProviderRepository    repositoryinterfaces.Repository[v1alpha1.MachineProvider]
)

func InitializeRepositories() {
	MachineProviderRepository = machineproviderrepository.NewMachineProviderRepository()
	KubernetesProviderRepository = kubernetesproviderrepository.NewKubernetesProviderRepository()
}
