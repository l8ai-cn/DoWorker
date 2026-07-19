package agentpod

import "github.com/anthropics/agentsmesh/backend/internal/service/agent"

func withCoordinator(coord PodCoordinatorForOrchestrator) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.PodCoordinator = coord }
}

func withBilling(b BillingServiceForOrchestrator) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.BillingService = b }
}

func withUserSvc(u UserServiceForOrchestrator) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.UserService = u }
}

func withRepoSvc(r RepositoryServiceForOrchestrator) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.RepoService = r }
}

func withTicketSvc(ts TicketServiceForOrchestrator) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.TicketService = ts }
}

func withRunnerSelector(rs RunnerSelectorForOrchestrator) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.RunnerSelector = rs }
}

func withAgentResolver(ar AgentResolverForOrchestrator) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.AgentResolver = ar }
}

func withModelResources(m ModelResourceResolver) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.ModelResources = m }
}

func withAgentConfigProvider(provider *mockAgentConfigProvider) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) {
		d.ConfigBuilder = agent.NewConfigBuilder(provider, noopBundleLoader{})
		d.AgentResolver = &mockAgentResolver{agentDef: provider.agentDef, err: provider.agentErr}
	}
}

func ptrStr(value string) *string { return &value }

func testModelResourceID() *int64 {
	id := int64(9)
	return &id
}
