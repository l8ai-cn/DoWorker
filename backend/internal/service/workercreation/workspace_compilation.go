package workercreation

import (
	"context"
	"fmt"
	"sort"

	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func (resolver *workspaceResolver) ResolveCompilationReferences(
	ctx context.Context,
	scope specservice.Scope,
	workerType slugkit.Slug,
	workspace specdomain.Workspace,
	secretRefs map[string]specdomain.SecretReference,
) (compilationReferences, error) {
	references, err := resolver.resolveWorkspaceReferences(
		ctx,
		scope,
		workerType,
		workspace,
	)
	if err != nil {
		return compilationReferences{}, err
	}
	fields := make([]string, 0, len(secretRefs))
	for field := range secretRefs {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	seenBundles := make(
		map[int64]struct{},
		len(references.EnvBundleNames)+len(fields),
	)
	for _, id := range workspace.EnvBundleIDs {
		seenBundles[int64(id)] = struct{}{}
	}
	for _, field := range fields {
		reference := secretRefs[field]
		if err := resolver.ResolveSecretReference(
			ctx,
			scope,
			workerType,
			field,
			reference,
		); err != nil {
			return compilationReferences{}, err
		}
		if _, exists := seenBundles[reference.ID]; exists {
			continue
		}
		bundle, err := resolver.resolveEnvBundle(
			ctx,
			scope,
			workerType,
			reference.ID,
		)
		if err != nil {
			return compilationReferences{}, err
		}
		if err := appendEnvBundleName(&references, bundle.Name); err != nil {
			return compilationReferences{}, err
		}
		seenBundles[reference.ID] = struct{}{}
	}
	return references, nil
}

func appendEnvBundleName(
	references *compilationReferences,
	name string,
) error {
	for _, existing := range references.EnvBundleNames {
		if existing == name {
			return fmt.Errorf(
				"%w: environment bundle name %q is ambiguous",
				specservice.ErrInvalidDraft,
				name,
			)
		}
	}
	references.EnvBundleNames = append(references.EnvBundleNames, name)
	return nil
}
