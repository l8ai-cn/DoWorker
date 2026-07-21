package workerdependencyartifact

import (
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
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
	for _, document := range expected {
		snapshot, exists := actual[document.ID]
		if !exists && document.Required {
			return fmt.Errorf(
				"worker definition config document %q is missing from artifact",
				document.ID,
			)
		}
		if exists && (snapshot.Format != document.Format ||
			snapshot.TargetPath != document.TargetPath) {
			return fmt.Errorf(
				"worker definition config document %q does not match artifact",
				document.ID,
			)
		}
	}
	declared := make(map[string]struct{}, len(expected))
	for _, document := range expected {
		declared[document.ID] = struct{}{}
	}
	for documentID := range actual {
		if _, exists := declared[documentID]; !exists {
			return fmt.Errorf(
				"worker definition config document %q is not declared",
				documentID,
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
				"secret reference field %q does not match worker definition",
				reference.Field,
			)
		}
		switch reference.OwnerScope {
		case envbundle.OwnerScopeOrg:
			if reference.OwnerID != scope.OrganizationID {
				return fmt.Errorf("organization secret owner does not match plan scope")
			}
		case envbundle.OwnerScopeUser:
			if reference.OwnerID != scope.ActorID {
				return fmt.Errorf("user secret owner does not match plan actor")
			}
		default:
			return fmt.Errorf("secret owner scope %q is invalid", reference.OwnerScope)
		}
	}
	return nil
}
