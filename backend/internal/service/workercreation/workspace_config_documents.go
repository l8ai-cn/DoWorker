package workercreation

import (
	"context"
	"fmt"

	"github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func (resolver *workspaceResolver) resolveConfigBundleNames(
	ctx context.Context,
	scope specservice.Scope,
	workerType slugkit.Slug,
	bindings []specdomain.ConfigDocumentBinding,
) ([]string, error) {
	definition, err := resolver.workerDefinition(workerType)
	if err != nil {
		return nil, err
	}
	declared := configDocumentsByID(definition.ConfigDocuments)
	if len(declared) == 0 {
		if len(bindings) != 0 {
			return nil, invalidWorkspaceReference(
				"Worker 配置文件",
				0,
				"worker type does not declare configuration documents",
				nil,
			)
		}
		return []string{}, nil
	}
	bundleNames := make([]string, 0, len(bindings))
	seenDocuments := make(map[string]struct{}, len(bindings))
	seenBundles := make(map[int64]struct{}, len(bindings))
	for _, binding := range bindings {
		documentID := binding.DocumentID
		if _, exists := declared[documentID]; !exists {
			return nil, fmt.Errorf(
				"%w: Worker 配置文件 document %q is not declared by the worker type",
				specservice.ErrInvalidDraft,
				documentID,
			)
		}
		if _, exists := seenDocuments[documentID]; exists {
			return nil, fmt.Errorf(
				"%w: Worker 配置文件 duplicates document %q",
				specservice.ErrInvalidDraft,
				documentID,
			)
		}
		if _, exists := seenBundles[binding.ConfigBundleID]; exists {
			return nil, fmt.Errorf(
				"%w: Worker 配置文件 duplicates bundle id %d",
				specservice.ErrInvalidDraft,
				binding.ConfigBundleID,
			)
		}
		bundle, err := resolver.resolveEnvBundle(
			ctx,
			scope,
			workerType,
			binding.ConfigBundleID,
		)
		if err != nil {
			return nil, err
		}
		if bundle.Kind != envbundle.KindConfig {
			return nil, invalidWorkspaceReference(
				"Worker 配置文件",
				bundle.ID,
				"bundle kind is not config",
				nil,
			)
		}
		seenDocuments[documentID] = struct{}{}
		seenBundles[binding.ConfigBundleID] = struct{}{}
		bundleNames = append(bundleNames, bundle.Name)
	}
	for documentID, document := range declared {
		if !document.Required {
			continue
		}
		if _, exists := seenDocuments[documentID]; !exists {
			return nil, fmt.Errorf(
				"%w: Worker 配置文件 is missing document %q",
				specservice.ErrInvalidDraft,
				documentID,
			)
		}
	}
	return bundleNames, nil
}

func configDocumentsByID(
	documents []workerdefinition.ConfigDocument,
) map[string]workerdefinition.ConfigDocument {
	byID := make(map[string]workerdefinition.ConfigDocument, len(documents))
	for _, document := range documents {
		byID[document.ID] = document
	}
	return byID
}
