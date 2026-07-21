package workercreation

import (
	"fmt"
	"sort"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdependencyartifact"
)

func buildSecretResolutions(
	scope control.Scope,
	refs ArtifactReferences,
	spec workerspec.Spec,
	definition workerdefinition.Definition,
	workspace *workspaceResolver,
) ([]workerdependencyartifact.SecretReferenceResolution, error) {
	fields := make([]string, 0, len(spec.TypeConfig.SecretRefs))
	for field := range spec.TypeConfig.SecretRefs {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	result := make([]workerdependencyartifact.SecretReferenceResolution, 0, len(fields))
	for _, field := range fields {
		reference := spec.TypeConfig.SecretRefs[field]
		bundle := workspace.envBundles[reference.ID]
		if bundle == nil {
			return nil, fmt.Errorf("WorkerTemplate artifact secret bundle %d was not resolved", reference.ID)
		}
		binding, found := credentialBindingForField(definition, field)
		if !found {
			return nil, fmt.Errorf("WorkerTemplate artifact secret field %q is not declared", field)
		}
		pin, err := referencePin(scope, refs.SecretBundles[field], reference.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, workerdependencyartifact.SecretReferenceResolution{
			ResourceResolution: pin, Field: field,
			BundleKey: binding.Target.Name, OwnerScope: bundle.OwnerScope,
			OwnerID: bundle.OwnerID,
		})
	}
	return result, nil
}
