package repositoryinterfaces

import (
	"context"
)

// Repository is a generic interface for all repository types
type Repository[T any] interface {
	// GetByUID retrieves a resource by its UID
	GetByUID(ctx context.Context, uid string) (T, error)

	// GetAll retrieves all resources of the given type
	GetAll(ctx context.Context) ([]T, error)

	// GetByName retrieves a resource by its name
	GetByName(ctx context.Context, name string) (T, error)
}
