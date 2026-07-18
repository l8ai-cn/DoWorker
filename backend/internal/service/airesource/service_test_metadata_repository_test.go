package airesource

import (
	"context"
	"errors"
	"slices"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
)

var errMissingAIResource = errors.New("missing AI resource")

func (r *memoryRepository) SaveConnectionMetadata(_ context.Context, value *domain.Connection) error {
	if err := r.failure("SaveConnectionMetadata"); err != nil {
		return err
	}
	stored := r.connections[value.ID]
	if stored == nil {
		return errMissingAIResource
	}
	if stored.Revision != value.Revision || !stored.UpdatedAt.Equal(value.UpdatedAt) {
		return domain.ErrConflict
	}
	if stored.OwnerScope != value.OwnerScope || stored.OwnerID != value.OwnerID ||
		stored.Identifier != value.Identifier || stored.ProviderKey != value.ProviderKey ||
		stored.BaseURL != value.BaseURL {
		return domain.ErrConflict
	}
	value.UpdatedAt = time.Now().UTC()
	copy := *value
	copy.ConfiguredFields = append([]string(nil), value.ConfiguredFields...)
	r.connections[value.ID] = &copy
	return nil
}

func (r *memoryRepository) SaveResourceMetadata(_ context.Context, value *domain.ModelResource) error {
	if err := r.failure("SaveResourceMetadata"); err != nil {
		return err
	}
	stored := r.resources[value.ID]
	if stored == nil {
		return errMissingAIResource
	}
	if stored.Revision != value.Revision || !stored.UpdatedAt.Equal(value.UpdatedAt) {
		return domain.ErrConflict
	}
	if stored.ProviderConnectionID != value.ProviderConnectionID ||
		stored.Identifier != value.Identifier || stored.ModelID != value.ModelID ||
		!slices.Equal(stored.Modalities, value.Modalities) ||
		!slices.Equal(stored.Capabilities, value.Capabilities) {
		return domain.ErrConflict
	}
	value.UpdatedAt = time.Now().UTC()
	r.resources[value.ID] = cloneResource(value)
	return nil
}
