package workerdependencyartifact

import (
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
)

func validateDefinitionConfigDocuments(
	expected []workerdefinition.ConfigDocument,
	bundles []workerdependency.RuntimeBundle,
) error {
	actual := make(map[string]workerdependency.ConfigDocument, len(expected))
	for _, bundle := range bundles {
		if bundle.ConfigDocument != nil {
			actual[bundle.ConfigDocument.ID] = *bundle.ConfigDocument
		}
	}
	if len(expected) != len(actual) {
		return fmt.Errorf("worker definition config documents do not match artifact")
	}
	for _, document := range expected {
		snapshot, exists := actual[document.ID]
		if !exists ||
			snapshot.Format != document.Format ||
			snapshot.TargetPath != document.TargetPath {
			return fmt.Errorf(
				"worker definition config document %q does not match artifact",
				document.ID,
			)
		}
	}
	return nil
}

func validateDefinitionSecrets(
	scope control.Scope,
	definition workerdefinition.Definition,
	references []workerdependency.SecretReference,
) error {
	bindings := make(map[string]workerdefinition.CredentialBinding)
	for _, binding := range definition.CredentialBindings {
		if binding.Source.Kind == "credential_bundle" {
			bindings[binding.Target.Name] = binding
		}
	}
	for _, reference := range references {
		binding, exists := bindings[reference.Field]
		if !exists || reference.BundleKey != binding.Target.Name {
			return fmt.Errorf(
				"Secret reference field %q does not match worker definition",
				reference.Field,
			)
		}
		switch reference.OwnerScope {
		case envbundle.OwnerScopeOrg:
			if reference.OwnerID != scope.OrganizationID {
				return fmt.Errorf("organization Secret owner does not match Plan scope")
			}
		case envbundle.OwnerScopeUser:
			if reference.OwnerID != scope.ActorID {
				return fmt.Errorf("user Secret owner does not match Plan actor")
			}
		default:
			return fmt.Errorf("Secret owner scope %q is invalid", reference.OwnerScope)
		}
	}
	return nil
}
