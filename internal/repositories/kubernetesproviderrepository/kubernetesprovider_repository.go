package kubernetesproviderrepository

import (
	"context"
	"encoding/json"

	"github.com/vitistack/vitistack-operator/internal/cache"
	"github.com/vitistack/vitistack-operator/internal/repositoryinterfaces"

	"github.com/vitistack/crds/pkg/v1alpha1"
)

// KubernetesProviderRepository interface defines operations for Kubernetes providers
type KubernetesProviderRepository interface {
	// Repository interface methods
	repositoryinterfaces.Repository[v1alpha1.KubernetesProvider]
}

// KubernetesProviderRepositoryImpl implements KubernetesProviderRepository
type KubernetesProviderRepositoryImpl struct {
}

func NewKubernetesProviderRepository() KubernetesProviderRepository {
	return &KubernetesProviderRepositoryImpl{}
}

// GetByUID implements Repository.GetByUID
func (m *KubernetesProviderRepositoryImpl) GetByUID(ctx context.Context, uid string) (v1alpha1.KubernetesProvider, error) {
	stringvalue, err := cache.Cache.Get(ctx, uid)
	if err != nil {
		return v1alpha1.KubernetesProvider{}, err
	}

	// Check if the string value is empty
	if stringvalue == "" {
		return v1alpha1.KubernetesProvider{}, nil
	}

	var KubernetesProvider v1alpha1.KubernetesProvider
	err = json.Unmarshal([]byte(stringvalue), &KubernetesProvider)
	if err != nil {
		return v1alpha1.KubernetesProvider{}, err
	}

	if KubernetesProvider.Kind != "KubernetesProvider" {
		return v1alpha1.KubernetesProvider{}, nil
	}

	return KubernetesProvider, nil
}

// GetAll implements Repository.GetAll
func (m *KubernetesProviderRepositoryImpl) GetAll(ctx context.Context) ([]v1alpha1.KubernetesProvider, error) {
	KubernetesProviderIds, err := cache.Cache.Keys(ctx)
	if err != nil {
		return nil, err
	}

	if len(KubernetesProviderIds) == 0 {
		return nil, nil
	}

	KubernetesProviders := make([]v1alpha1.KubernetesProvider, 0)
	for _, KubernetesProviderId := range KubernetesProviderIds {
		KubernetesProviderString, err := cache.Cache.Get(ctx, KubernetesProviderId)
		if err != nil {
			continue
		}

		var KubernetesProvider v1alpha1.KubernetesProvider
		err = json.Unmarshal([]byte(KubernetesProviderString), &KubernetesProvider)
		if err != nil {
			continue
		}

		if KubernetesProvider.Kind != "KubernetesProvider" {
			continue
		}

		KubernetesProviders = append(KubernetesProviders, KubernetesProvider)
	}
	return KubernetesProviders, nil
}

// GetByName implements Repository.GetByName
func (m *KubernetesProviderRepositoryImpl) GetByName(ctx context.Context, name string) (v1alpha1.KubernetesProvider, error) {
	KubernetesProviderIds, err := cache.Cache.Keys(ctx)
	if err != nil {
		return v1alpha1.KubernetesProvider{}, err
	}

	if len(KubernetesProviderIds) == 0 {
		return v1alpha1.KubernetesProvider{}, nil
	}

	for _, KubernetesProviderId := range KubernetesProviderIds {
		KubernetesProviderString, err := cache.Cache.Get(ctx, KubernetesProviderId)
		if err != nil {
			continue
		}

		var KubernetesProvider v1alpha1.KubernetesProvider
		err = json.Unmarshal([]byte(KubernetesProviderString), &KubernetesProvider)
		if err != nil {
			continue
		}

		if KubernetesProvider.Kind != "KubernetesProvider" {
			continue
		}

		if KubernetesProvider.Name == name {
			return KubernetesProvider, nil
		}
	}
	return v1alpha1.KubernetesProvider{}, nil
}
