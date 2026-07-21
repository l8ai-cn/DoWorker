package airesource

import (
	"context"
	"fmt"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/audit"
)

func (s *Service) CreateResource(ctx context.Context, actor Actor, input CreateResourceInput) (ResourceView, error) {
	connection, _, err := s.connectionForActor(ctx, actor, input.ConnectionID, true)
	if err != nil {
		return ResourceView{}, err
	}
	provider, exists := domain.Provider(connection.ProviderKey.String())
	if !exists {
		return ResourceView{}, ErrInvalidProvider
	}
	if err := validateResourceDefinition(provider, input.Modalities, input.Capabilities); err != nil {
		return ResourceView{}, err
	}
	resource := &domain.ModelResource{
		ProviderConnectionID: connection.ID, Identifier: input.Identifier, ModelID: input.ModelID,
		DisplayName: input.DisplayName, Modalities: append([]domain.Modality(nil), input.Modalities...),
		Capabilities: append([]domain.Capability(nil), input.Capabilities...), Status: connection.Status, IsEnabled: true,
	}
	if err := domain.ValidateModelResource(*resource); err != nil {
		return ResourceView{}, fmt.Errorf("%w: invalid model resource: %v", ErrInvalidRequirements, err)
	}
	err = s.mutations.Run(ctx, func(repo domain.Repository, recorder AuditRecorder) error {
		if createErr := repo.CreateResource(ctx, resource); createErr != nil {
			return createErr
		}
		return recordAudit(ctx, recorder, actor, audit.ActionModelResourceCreated, audit.ResourceModelResource, resource.ID, connection, "success", audit.Details{"resource_identifier": resource.Identifier.String()})
	})
	if err != nil {
		return ResourceView{}, err
	}
	view := resourceView(resource)
	return view, nil
}

func (s *Service) UpdateResource(ctx context.Context, actor Actor, resourceID int64, input UpdateResourceInput) (ResourceView, error) {
	resource, connection, _, err := s.resourceForActor(ctx, actor, resourceID, true)
	if err != nil {
		return ResourceView{}, err
	}
	provider, exists := domain.Provider(connection.ProviderKey.String())
	if !exists {
		return ResourceView{}, ErrInvalidProvider
	}
	if err := validateResourceDefinition(provider, input.Modalities, input.Capabilities); err != nil {
		return ResourceView{}, err
	}
	candidate := *resource
	candidate.ModelID, candidate.DisplayName = input.ModelID, input.DisplayName
	candidate.Modalities = append([]domain.Modality(nil), input.Modalities...)
	candidate.Capabilities = append([]domain.Capability(nil), input.Capabilities...)
	if err := domain.ValidateModelResource(candidate); err != nil {
		return ResourceView{}, fmt.Errorf("%w: invalid model resource: %v", ErrInvalidRequirements, err)
	}
	runtimeChanged := resourceRuntimeChanged(resource, &candidate)
	resource.DisplayName = candidate.DisplayName
	if runtimeChanged {
		resource.ModelID = candidate.ModelID
		resource.Modalities = candidate.Modalities
		resource.Capabilities = candidate.Capabilities
	}
	resource.Status = connection.Status
	resource.LastValidatedAt = connection.LastValidatedAt
	resource.ValidationError = connection.ValidationError
	err = s.mutations.Run(ctx, func(repo domain.Repository, recorder AuditRecorder) error {
		save := repo.SaveResourceMetadata
		if runtimeChanged {
			save = repo.SaveResource
		}
		if saveErr := save(ctx, resource); saveErr != nil {
			return saveErr
		}
		return recordAudit(ctx, recorder, actor, audit.ActionModelResourceUpdated, audit.ResourceModelResource, resource.ID, connection, "success", audit.Details{"resource_identifier": resource.Identifier.String()})
	})
	if err != nil {
		return ResourceView{}, err
	}
	view := resourceView(resource)
	return view, nil
}

func (s *Service) SetResourceEnabled(ctx context.Context, actor Actor, resourceID int64, enabled bool) error {
	resource, connection, _, err := s.resourceForActor(ctx, actor, resourceID, true)
	if err != nil {
		return err
	}
	resource.IsEnabled = enabled
	action := audit.ActionModelResourceDisabled
	if enabled {
		action = audit.ActionModelResourceEnabled
	}
	return s.mutations.Run(ctx, func(repo domain.Repository, recorder AuditRecorder) error {
		if saveErr := repo.SaveResourceMetadata(ctx, resource); saveErr != nil {
			return saveErr
		}
		return recordAudit(ctx, recorder, actor, action, audit.ResourceModelResource, resource.ID, connection, "success", audit.Details{"resource_identifier": resource.Identifier.String()})
	})
}

func (s *Service) DeleteResource(ctx context.Context, actor Actor, resourceID int64) error {
	resource, connection, _, err := s.resourceForActor(ctx, actor, resourceID, true)
	if err != nil {
		return err
	}
	return s.mutations.Run(ctx, func(repo domain.Repository, recorder AuditRecorder) error {
		if deleteErr := repo.DeleteResource(ctx, resourceID, resource.Revision, resource.UpdatedAt); deleteErr != nil {
			return deleteErr
		}
		return recordAudit(ctx, recorder, actor, audit.ActionModelResourceDeleted, audit.ResourceModelResource, resource.ID, connection, "success", audit.Details{"resource_identifier": resource.Identifier.String()})
	})
}

func (s *Service) SetDefault(ctx context.Context, actor Actor, resourceID int64, modality domain.Modality) error {
	resource, connection, _, err := s.resourceForActor(ctx, actor, resourceID, true)
	if err != nil {
		return err
	}
	if !supportsModality(resource.Modalities, modality) {
		return ErrIncompatibleModality
	}
	return s.mutations.Run(ctx, func(repo domain.Repository, recorder AuditRecorder) error {
		if defaultErr := repo.SetDefault(ctx, resourceID, modality); defaultErr != nil {
			return defaultErr
		}
		return recordAudit(ctx, recorder, actor, audit.ActionModelResourceDefaulted, audit.ResourceModelResource, resource.ID, connection, "success", audit.Details{"modality": string(modality), "resource_identifier": resource.Identifier.String()})
	})
}

func validateResourceDefinition(provider domain.ProviderDefinition, modalities []domain.Modality, capabilities []domain.Capability) error {
	for _, modality := range modalities {
		if !modality.Valid() || !supportsModality(provider.Modalities, modality) {
			return fmt.Errorf("%w: %s", ErrIncompatibleModality, modality)
		}
		if !supportsCapability(capabilities, modality) {
			return fmt.Errorf("%w: capability for %s", ErrIncompatibleModality, modality)
		}
	}
	return nil
}

func supportsCapability(capabilities []domain.Capability, modality domain.Modality) bool {
	wanted := map[domain.Capability]bool{}
	switch modality {
	case domain.ModalityChat:
		wanted[domain.CapabilityTextGeneration] = true
	case domain.ModalityImage:
		wanted[domain.CapabilityImageGeneration] = true
	case domain.ModalityAudio:
		wanted[domain.CapabilitySpeechToText], wanted[domain.CapabilityTextToSpeech] = true, true
	case domain.ModalityVideo:
		wanted[domain.CapabilityVideoGeneration] = true
	case domain.ModalityEmbedding:
		wanted[domain.CapabilityEmbedding] = true
	case domain.ModalityMultimodal:
		wanted[domain.CapabilityVisionInput], wanted[domain.CapabilityTextGeneration] = true, true
	}
	for _, capability := range capabilities {
		if wanted[capability] {
			return true
		}
	}
	return false
}

func supportsModality(modalities []domain.Modality, required domain.Modality) bool {
	for _, modality := range modalities {
		if modality == required {
			return true
		}
	}
	return false
}
