package workercreation

import (
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdependencyartifact"
)

func configDocument(
	definition workerdefinition.Definition,
	id string,
) (*workerdependencyartifact.ConfigDocumentResolution, error) {
	for _, document := range definition.ConfigDocuments {
		if document.ID != id {
			continue
		}
		return &workerdependencyartifact.ConfigDocumentResolution{
			ID: document.ID, Format: document.Format,
			TargetPath: document.TargetPath,
		}, nil
	}
	return nil, fmt.Errorf("WorkerTemplate artifact config document %q is not declared", id)
}

func materializedRuntimeValues(
	values []workerdependencyartifact.RuntimeValueResolution,
) []workerdependency.RuntimeValue {
	result := make([]workerdependency.RuntimeValue, len(values))
	for index, value := range values {
		result[index] = workerdependency.RuntimeValue{
			Name: value.Name, Value: value.Value,
		}
	}
	return result
}
