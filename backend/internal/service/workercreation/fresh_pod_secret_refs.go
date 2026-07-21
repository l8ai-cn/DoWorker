package workercreation

import (
	"context"
	"fmt"
	"sort"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type effectiveEnvBundleLookup interface {
	ListEffectiveForUser(
		context.Context,
		int64,
		int64,
		string,
	) ([]*envbundle.EnvBundle, error)
}

func (service *Service) freshPodSecretRefs(
	ctx context.Context,
	scope specservice.Scope,
	workerType slugkit.Slug,
	schema specdomain.TypeSchema,
) (map[string]specdomain.SecretReference, error) {
	definition, err := service.freshPodDefinition(workerType)
	if err != nil {
		return nil, err
	}
	fields := requiredCredentialBundleFields(definition, schema)
	refs := make(map[string]specdomain.SecretReference, len(fields))
	if len(fields) == 0 {
		return refs, nil
	}
	lookup, ok := service.workspaceDeps.EnvBundles.(effectiveEnvBundleLookup)
	if !ok {
		return nil, specservice.ErrResolverUnavailable
	}
	bundles, err := lookup.ListEffectiveForUser(
		ctx,
		scope.UserID,
		scope.OrgID,
		workerType.String(),
	)
	if err != nil {
		return nil, err
	}
	kind, err := slugkit.NewFromTrusted("env-bundle")
	if err != nil {
		return nil, err
	}
	for _, field := range fields {
		binding, _ := credentialBindingForField(definition, field)
		bundle, err := selectFreshCredentialBundle(
			scope,
			workerType,
			binding,
			bundles,
		)
		if err != nil {
			return nil, err
		}
		refs[field] = specdomain.SecretReference{Kind: kind, ID: bundle.ID}
	}
	return refs, nil
}

func (service *Service) freshPodDefinition(
	workerType slugkit.Slug,
) (workerdefinition.Definition, error) {
	if service == nil || service.workspaceDeps.Definitions == nil {
		return workerdefinition.Definition{}, specservice.ErrResolverUnavailable
	}
	definition, ok := service.workspaceDeps.Definitions.Get(workerType.String())
	if !ok {
		return workerdefinition.Definition{}, invalidWorkerType("missing canonical definition")
	}
	return definition, nil
}

func requiredCredentialBundleFields(
	definition workerdefinition.Definition,
	schema specdomain.TypeSchema,
) []string {
	fields := make([]string, 0, len(schema.Fields))
	for field, fieldSchema := range schema.Fields {
		if fieldSchema.Kind != specdomain.TypeFieldSecret || !fieldSchema.Required {
			continue
		}
		binding, found := credentialBindingForField(definition, field)
		if found && binding.Source.Kind == "credential_bundle" {
			fields = append(fields, field)
		}
	}
	sort.Strings(fields)
	return fields
}

func selectFreshCredentialBundle(
	scope specservice.Scope,
	workerType slugkit.Slug,
	binding workerdefinition.CredentialBinding,
	bundles []*envbundle.EnvBundle,
) (*envbundle.EnvBundle, error) {
	var selected *envbundle.EnvBundle
	selectedScore := -1
	tied := false
	for _, bundle := range bundles {
		if !freshCredentialBundleMatches(scope, workerType, binding, bundle) {
			continue
		}
		score := freshCredentialBundleScore(bundle)
		switch {
		case score > selectedScore:
			selected = bundle
			selectedScore = score
			tied = false
		case score == selectedScore:
			tied = true
		}
	}
	if selected == nil {
		return nil, invalidFreshPodDraft(
			binding.Target.Name,
			fmt.Sprintf("credential bundle %q is required", binding.Source.Ref),
		)
	}
	if tied {
		return nil, invalidFreshPodDraft(
			binding.Target.Name,
			fmt.Sprintf("credential bundle %q is ambiguous", binding.Source.Ref),
		)
	}
	return selected, nil
}

func freshCredentialBundleMatches(
	scope specservice.Scope,
	workerType slugkit.Slug,
	binding workerdefinition.CredentialBinding,
	bundle *envbundle.EnvBundle,
) bool {
	if bundle == nil || !bundle.IsActive || bundle.Kind != envbundle.KindCredential {
		return false
	}
	if bundle.Name != binding.Source.Ref || !envBundleVisibleTo(bundle, scope) {
		return false
	}
	return bundle.AgentSlug == nil || *bundle.AgentSlug == workerType.String()
}

func freshCredentialBundleScore(bundle *envbundle.EnvBundle) int {
	score := 0
	if bundle.OwnerScope == envbundle.OwnerScopeUser {
		score += 10
	}
	if bundle.AgentSlug != nil {
		score += 5
	}
	return score
}
