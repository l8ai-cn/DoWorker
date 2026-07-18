package orchestrationcontrol

import "github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"

type workerDefinitionPolicyStub map[string]workerdefinition.EnvironmentBundlePolicy

func (stub workerDefinitionPolicyStub) EnvironmentBundlePolicy(
	workerType string,
) (workerdefinition.EnvironmentBundlePolicy, bool) {
	policy, found := stub[workerType]
	return policy, found
}
