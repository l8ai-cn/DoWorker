package workercreation

import (
	"context"
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

func (resolver *workspaceResolver) resolveConfigDocumentIDs(
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
	if len(bindings) > len(declared) {
		return nil, fmt.Errorf(
			"%w: Worker 配置文件 has more bindings than declared documents",
			specservice.ErrInvalidDraft,
		)
	}
	documentIDs := make([]string, 0, len(bindings))
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
		documentIDs = append(documentIDs, documentID)
	}
	for documentID, document := range declared {
		if document.Required {
			if _, exists := seenDocuments[documentID]; exists {
				continue
			}
			return nil, fmt.Errorf(
				"%w: Worker 配置文件 is missing document %q",
				specservice.ErrInvalidDraft,
				documentID,
			)
		}
	}
	return documentIDs, nil
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
