package workerdefinition

import "sort"

type EnvironmentBundlePolicy struct {
	ModelManagedFields     []string
	CredentialBundleFields []string
}

func BuildEnvironmentBundlePolicy(
	definition Definition,
) EnvironmentBundlePolicy {
	modelManaged := make(map[string]struct{}, len(definition.CredentialBindings)+1)
	credentialBundle := make(map[string]struct{}, len(definition.CredentialBindings))
	if definition.ModelRequirement.Required {
		modelManaged["model"] = struct{}{}
	}
	for _, binding := range definition.CredentialBindings {
		switch binding.Source.Kind {
		case "model_resource":
			modelManaged[binding.Target.Name] = struct{}{}
		case "credential_bundle":
			credentialBundle[binding.Target.Name] = struct{}{}
		}
	}
	return EnvironmentBundlePolicy{
		ModelManagedFields:     sortedEnvironmentFields(modelManaged),
		CredentialBundleFields: sortedEnvironmentFields(credentialBundle),
	}
}

func sortedEnvironmentFields(fields map[string]struct{}) []string {
	result := make([]string, 0, len(fields))
	for field := range fields {
		result = append(result, field)
	}
	sort.Strings(result)
	return result
}
