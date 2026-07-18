package workerdependency

import (
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

func validateSecretReferences(
	document Document,
	bundles []RuntimeBundle,
	references []SecretReference,
) error {
	materialized := make(map[int64]struct{}, len(bundles))
	for _, bundle := range bundles {
		materialized[bundle.Pin.DomainID] = struct{}{}
	}
	fields := make(map[string]struct{}, len(references))
	modelFields := modelManagedEnvironmentFields(document)
	credentialFields := make(
		map[string]struct{},
		len(document.Worker.CredentialBundleFields),
	)
	for _, field := range document.Worker.CredentialBundleFields {
		credentialFields[field] = struct{}{}
	}
	for _, reference := range references {
		if err := validatePin(
			document,
			reference.Pin,
			resource.KindEnvironmentBundle,
		); err != nil {
			return err
		}
		if _, exists := materialized[reference.Pin.DomainID]; exists {
			return fmt.Errorf("secret EnvironmentBundle cannot be materialized")
		}
		if err := workerspec.ValidateConfigField(reference.Field); err != nil {
			return fmt.Errorf("secret target field: %w", err)
		}
		if _, exists := modelFields[reference.Field]; exists {
			return fmt.Errorf(
				"secret target field %q is managed by a model resource",
				reference.Field,
			)
		}
		if _, exists := credentialFields[reference.Field]; !exists {
			return fmt.Errorf(
				"secret target field %q is not owned by a credential bundle",
				reference.Field,
			)
		}
		if err := workerspec.ValidateConfigField(reference.BundleKey); err != nil {
			return fmt.Errorf("secret bundle key: %w", err)
		}
		if _, exists := fields[reference.Field]; exists {
			return fmt.Errorf("duplicate secret target field %q", reference.Field)
		}
		fields[reference.Field] = struct{}{}
		if reference.OwnerID <= 0 {
			return fmt.Errorf("secret owner id must be positive")
		}
		switch reference.OwnerScope {
		case envbundle.OwnerScopeUser:
		case envbundle.OwnerScopeOrg:
			if reference.OwnerID != document.OrganizationID {
				return fmt.Errorf("organization secret owner does not match artifact")
			}
		default:
			return fmt.Errorf("secret owner scope %q is invalid", reference.OwnerScope)
		}
	}
	return nil
}

func modelManagedEnvironmentFields(document Document) map[string]struct{} {
	fields := make(map[string]struct{}, len(document.Worker.ModelManagedFields)+3)
	for _, field := range document.Worker.ModelManagedFields {
		fields[field] = struct{}{}
	}
	for _, tool := range document.Models.Tools {
		fields[tool.Environment.APIKeyTarget] = struct{}{}
		fields[tool.Environment.BaseURLTarget] = struct{}{}
		fields[tool.Environment.ModelIDTarget] = struct{}{}
	}
	return fields
}
