package workercreation

import "github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"

func (service *Service) EnvironmentBundlePolicy(
	workerType string,
) (workerdefinition.EnvironmentBundlePolicy, bool) {
	if service == nil || service.workspaceDeps.Definitions == nil {
		return workerdefinition.EnvironmentBundlePolicy{}, false
	}
	definition, found := service.workspaceDeps.Definitions.Get(workerType)
	if !found {
		return workerdefinition.EnvironmentBundlePolicy{}, false
	}
	return workerdefinition.BuildEnvironmentBundlePolicy(definition), true
}
