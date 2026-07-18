package orchestrationcontrol

import (
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
)

func (service *Service) applyWorkerDefinitionPolicy(
	filter ResourceListFilter,
) (ResourceListFilter, error) {
	if filter.EnvironmentBundle == nil {
		return filter, nil
	}
	policy, found := service.workerDefinitions.EnvironmentBundlePolicy(
		filter.EnvironmentBundle.WorkerType.String(),
	)
	if !found {
		return ResourceListFilter{}, fmt.Errorf(
			"%w: Worker definition does not exist",
			control.ErrInvalid,
		)
	}
	resolved := *filter.EnvironmentBundle
	switch resolved.Purpose {
	case EnvironmentBundlePurposeRuntime:
		resolved.ModelManagedFields = append(
			[]string{},
			policy.ModelManagedFields...,
		)
	case EnvironmentBundlePurposeCredential:
		if !containsField(policy.CredentialBundleFields, resolved.TargetName) {
			return ResourceListFilter{}, invalidEnvironmentBundleFilter()
		}
	}
	filter.EnvironmentBundle = &resolved
	return filter, nil
}

func containsField(fields []string, target string) bool {
	for _, field := range fields {
		if field == target {
			return true
		}
	}
	return false
}
