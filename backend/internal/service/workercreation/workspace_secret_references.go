package workercreation

import (
	"context"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func (resolver *workspaceResolver) ResolveSecretReference(
	ctx context.Context,
	scope specservice.Scope,
	workerType slugkit.Slug,
	field string,
	reference specdomain.SecretReference,
) error {
	if reference.Kind.String() != "env-bundle" {
		return invalidWorkspaceReference(field, reference.ID, "unsupported secret reference kind", nil)
	}
	definition, err := resolver.workerDefinition(workerType)
	if err != nil {
		return err
	}
	binding, found := credentialBindingForField(definition, field)
	if !found {
		return invalidWorkspaceReference(field, reference.ID, "credential field is not declared by the worker definition", nil)
	}
	if binding.Source.Kind != "credential_bundle" {
		return invalidWorkspaceReference(field, reference.ID, "credential field is managed by the model resource", nil)
	}
	bundle, err := resolver.resolveEnvBundle(ctx, scope, workerType, reference.ID)
	if err != nil {
		return err
	}
	if bundle.Kind != envbundle.KindCredential {
		return invalidWorkspaceReference(field, reference.ID, "secret reference requires a credential bundle", nil)
	}
	if _, exists := bundle.Data[binding.Target.Name]; !exists {
		return invalidWorkspaceReference(field, reference.ID, "credential bundle does not configure the declared field", nil)
	}
	return nil
}

func credentialBindingForField(
	definition workerdefinition.Definition,
	field string,
) (workerdefinition.CredentialBinding, bool) {
	for _, binding := range definition.CredentialBindings {
		if binding.Target.Name == field {
			return binding, true
		}
	}
	return workerdefinition.CredentialBinding{}, false
}
