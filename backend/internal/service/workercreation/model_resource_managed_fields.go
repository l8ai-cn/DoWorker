package workercreation

import (
	"sort"

	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
)

func modelResourceManagedFields(
	definition workerdefinition.Definition,
) map[string]struct{} {
	fields := make(map[string]struct{}, len(definition.CredentialBindings)+1)
	if definition.ModelRequirement.Required {
		fields["model"] = struct{}{}
	}
	for _, binding := range definition.CredentialBindings {
		if binding.Source.Kind == "model_resource" {
			fields[binding.Target.Name] = struct{}{}
		}
	}
	return fields
}

func isModelResourceManagedField(
	definition workerdefinition.Definition,
	field string,
) bool {
	_, exists := modelResourceManagedFields(definition)[field]
	return exists
}

func modelResourceManagedRuntimeField(
	definition workerdefinition.Definition,
	values map[string]string,
) string {
	fields := make([]string, 0, len(values))
	for field := range values {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	for _, field := range fields {
		if isModelResourceManagedField(definition, field) {
			return field
		}
	}
	return ""
}
