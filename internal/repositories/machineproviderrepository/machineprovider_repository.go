package machineproviderrepository

import (
	"context"
	"encoding/json"

	"github.com/NorskHelsenett/oss-datacenter-operator/internal/cache"
	"github.com/NorskHelsenett/oss-datacenter-operator/internal/repositoryinterfaces"

	"github.com/vitistack/crds/pkg/v1alpha1"
)

// MachineProviderRepository interface defines operations for Machine providers
type MachineProviderRepository interface {
	// Repository interface methods
	repositoryinterfaces.Repository[v1alpha1.MachineProvider]
}

// MachineProviderRepositoryImpl implements MachineProviderRepository
type MachineProviderRepositoryImpl struct {
}

func NewMachineProviderRepository() MachineProviderRepository {
	return &MachineProviderRepositoryImpl{}
}

// GetByUID implements Repository.GetByUID
func (m *MachineProviderRepositoryImpl) GetByUID(ctx context.Context, uid string) (v1alpha1.MachineProvider, error) {
	stringvalue, err := cache.Cache.Get(ctx, uid)
	if err != nil {
		return v1alpha1.MachineProvider{}, err
	}

	// Check if the string value is empty
	if stringvalue == "" {
		return v1alpha1.MachineProvider{}, nil
	}

	var machineProvider v1alpha1.MachineProvider
	err = json.Unmarshal([]byte(stringvalue), &machineProvider)
	if err != nil {
		return v1alpha1.MachineProvider{}, err
	}

	if machineProvider.Kind != "MachineProvider" {
		return v1alpha1.MachineProvider{}, nil
	}

	return machineProvider, nil
}

// GetAll implements Repository.GetAll
func (m *MachineProviderRepositoryImpl) GetAll(ctx context.Context) ([]v1alpha1.MachineProvider, error) {
	machineProviderIds, err := cache.Cache.Keys(ctx)
	if err != nil {
		return nil, err
	}

	if len(machineProviderIds) == 0 {
		return nil, nil
	}

	machineProviders := make([]v1alpha1.MachineProvider, 0)
	for _, machineProviderId := range machineProviderIds {
		machineProviderString, err := cache.Cache.Get(ctx, machineProviderId)
		if err != nil {
			continue
		}

		var machineProvider v1alpha1.MachineProvider
		err = json.Unmarshal([]byte(machineProviderString), &machineProvider)
		if err != nil {
			continue
		}

		if machineProvider.Kind != "MachineProvider" {
			continue
		}

		machineProviders = append(machineProviders, machineProvider)
	}
	return machineProviders, nil
}

// GetByName implements Repository.GetByName
func (m *MachineProviderRepositoryImpl) GetByName(ctx context.Context, name string) (v1alpha1.MachineProvider, error) {
	machineProviderIds, err := cache.Cache.Keys(ctx)
	if err != nil {
		return v1alpha1.MachineProvider{}, err
	}

	if len(machineProviderIds) == 0 {
		return v1alpha1.MachineProvider{}, nil
	}

	for _, machineProviderId := range machineProviderIds {
		machineProviderString, err := cache.Cache.Get(ctx, machineProviderId)
		if err != nil {
			continue
		}

		var machineProvider v1alpha1.MachineProvider
		err = json.Unmarshal([]byte(machineProviderString), &machineProvider)
		if err != nil {
			continue
		}

		if machineProvider.Kind != "MachineProvider" {
			continue
		}

		if machineProvider.Name == name {
			return machineProvider, nil
		}
	}
	return v1alpha1.MachineProvider{}, nil
}
