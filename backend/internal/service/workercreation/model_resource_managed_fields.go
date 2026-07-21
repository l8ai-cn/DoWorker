package workercreation

import (
	"sort"

	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
)

func modelResourceManagedFields(
	definition workerdefinition.Definition,
) map[string]struct{} {
	policy := workerdefinition.BuildEnvironmentBundlePolicy(definition)
	fields := make(map[string]struct{}, len(policy.ModelManagedFields))
	for _, field := range policy.ModelManagedFields {
		fields[field] = struct{}{}
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
