package agentpod

import (
	"context"
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/agentfile"
	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	resourcesvc "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func workerSpecPlanRequestForTest(
	t *testing.T,
	ctx context.Context,
	agentSlug string,
	agentfileLayer string,
	resource *resourcesvc.ResolvedResource,
) (*OrchestrateCreatePodRequest, *workerCreationPreparer) {
	t.Helper()
	spec := workerSpecForPlanSource(agentSlug, agentfileLayer, resource)
	draft := workercreation.Draft{
		OptionsRevision: "test-options",
		WorkerSpec:      workerSpecDraftFromSpec(spec),
	}
	prepared := workercreation.Prepared{
		Snapshot:       resolvedWorkerSpecFromSpecForPodServiceTest(t, ctxOrgID(ctx), spec),
		Spec:           spec,
		AgentfileLayer: normalizedPlanAgentfileLayer(agentfileLayer),
	}
	req := &OrchestrateCreatePodRequest{
		OrganizationID:  ctxOrgID(ctx),
		UserID:          ctxUserID(ctx),
		WorkerSpecDraft: &draft,
	}
	return req, &workerCreationPreparer{prepared: prepared}
}

func createPodWithPlanSourceForTest(
	t *testing.T,
	orchestrator *PodOrchestrator,
	ctx context.Context,
	req *OrchestrateCreatePodRequest,
) (*OrchestrateCreatePodResult, error) {
	t.Helper()
	materializePlanSourceForTest(t, orchestrator, req)
	return (orchestrator).CreatePod(ctx, req)
}

func materializePlanSourceForTest(
	t *testing.T,
	orchestrator *PodOrchestrator,
	req *OrchestrateCreatePodRequest,
) {
	t.Helper()
	if req.WorkerSpecDraft != nil ||
		req.WorkerSpecSnapshotID != nil ||
		req.SourcePodKey != "" ||
		req.AgentSlug == "" {
		return
	}
	layer := workerSpecStringValue(req.AgentfileLayer)
	layer = appendKnowledgeLayerForTest(layer, req.KnowledgeMounts)
	if req.BranchName != nil &&
		req.RepositoryID == nil &&
		!strings.Contains(layer, "BRANCH") {
		layer = appendPlanLayerLineForTest(
			layer,
			"BRANCH "+agentfile.FormatStringLiteral(*req.BranchName),
		)
	}
	resource := planSourceModelResourceForTest(orchestrator, req.AgentSlug)
	spec := workerSpecForPlanSource(req.AgentSlug, layer, resource)
	spec.TypeConfig.InteractionMode = planSourceInteractionMode(
		req.AgentSlug,
		layer,
		orchestrator,
	)
	applyPlanSourceOverridesForTest(&spec, req)
	repository, prepareErr := planSourceRepositoryForTest(orchestrator, req)
	if prepareErr != nil {
		spec.Workspace.RepositoryID = nil
		spec.Workspace.Branch = ""
	}
	if prepareErr == nil && repository == nil && req.RepositoryID != nil {
		spec.Workspace.RepositoryID = nil
		spec.Workspace.Branch = ""
	}
	if repository != nil && strings.TrimSpace(spec.Workspace.Branch) == "" {
		spec.Workspace.Branch = firstNonEmpty(repository.DefaultBranch, "main")
	}
	req.WorkerSpecDraft = &workercreation.Draft{
		OptionsRevision: "test-options",
		WorkerSpec:      workerSpecDraftFromSpec(spec),
	}
	prepared := workercreation.Prepared{
		Snapshot:       resolvedWorkerSpecFromSpecForPodServiceTest(t, req.OrganizationID, spec),
		Spec:           spec,
		AgentfileLayer: normalizedPlanAgentfileLayer(layer),
		Repository:     repository,
	}
	orchestrator.workerCreation = &workerCreationPreparer{
		prepared: prepared,
		err:      prepareErr,
	}
	clearLegacyPlanFixtureFields(req)
}

func workerSpecForPlanSource(
	agentSlug string,
	agentfileLayer string,
	resource *resourcesvc.ResolvedResource,
) specdomain.Spec {
	spec := podServiceWorkerSpec()
	spec.Runtime.WorkerType.Slug = slugkit.MustNewForTest(agentSlug)
	spec.Runtime.ModelBinding = modelBindingFromResolvedResource(resource)
	spec.TypeConfig.InteractionMode = planSourceInteractionMode(agentSlug, agentfileLayer, nil)
	spec.Workspace.InitialTask = ""
	spec.Metadata.Alias = ""
	return spec
}

func applyPlanSourceOverridesForTest(
	spec *specdomain.Spec,
	req *OrchestrateCreatePodRequest,
) {
	if req.RepositoryID != nil {
		spec.Workspace.RepositoryID = cloneWorkerSpecInt64Pointer(req.RepositoryID)
	}
	if req.BranchName != nil && req.RepositoryID != nil {
		spec.Workspace.Branch = strings.TrimSpace(*req.BranchName)
	}
	if req.Alias != nil {
		spec.Metadata.Alias = strings.TrimSpace(*req.Alias)
	}
	if req.AutomationLevel != "" {
		spec.TypeConfig.AutomationLevel = specdomain.AutomationLevel(req.AutomationLevel)
	}
	if req.Perpetual {
		spec.Lifecycle.TerminationPolicy = specdomain.TerminationPolicyManual
	}
}

func workerSpecDraftFromSpec(spec specdomain.Spec) specservice.Draft {
	return specservice.Draft{
		ModelResourceID: spec.Runtime.ModelBinding.ResourceID,
		WorkerTypeSlug:  spec.Runtime.WorkerType.Slug,
		Runtime: specservice.RuntimeSelection{
			RuntimeImageID:    spec.Runtime.Image.ID,
			PlacementPolicy:   spec.Placement.Policy,
			ComputeTargetID:   spec.Placement.ComputeTarget.ID,
			DeploymentMode:    spec.Placement.DeploymentMode,
			ResourceProfileID: spec.Placement.ResourceProfile.ID,
		},
		TypeConfig: spec.TypeConfig,
		Workspace:  spec.Workspace,
		Lifecycle:  spec.Lifecycle,
		Metadata:   spec.Metadata,
	}
}

func modelBindingFromResolvedResource(resource *resourcesvc.ResolvedResource) specdomain.ModelBinding {
	return specdomain.ModelBinding{
		ResourceID:         resource.Resource.ID,
		ResourceRevision:   resource.Resource.Revision,
		ConnectionID:       resource.Connection.ID,
		ConnectionRevision: resource.Connection.Revision,
		ProviderKey:        resource.Connection.ProviderKey,
		ProtocolAdapter:    slugkit.Slug(resource.Provider.ProtocolAdapter),
		ModelID:            strings.TrimSpace(resource.Resource.ModelID),
	}
}

func planSourceInteractionMode(
	agentSlug string,
	agentfileLayer string,
	orchestrator *PodOrchestrator,
) specdomain.InteractionMode {
	if strings.Contains(agentfileLayer, "MODE acp") {
		return specdomain.InteractionModeACP
	}
	if strings.Contains(agentfileLayer, "MODE pty") {
		return specdomain.InteractionModePTY
	}
	if agentSlug == "codex-cli" && planSourceAgentSupportsACP(orchestrator) {
		return specdomain.InteractionModeACP
	}
	return specdomain.InteractionModePTY
}

func planSourceAgentSupportsACP(orchestrator *PodOrchestrator) bool {
	if orchestrator == nil {
		return false
	}
	if resolver, ok := orchestrator.agentResolver.(*mockAgentResolver); ok &&
		resolver.agentDef != nil {
		return strings.Contains(resolver.agentDef.SupportedModes, "acp")
	}
	return false
}

func normalizedPlanAgentfileLayer(agentfileLayer string) string {
	if strings.TrimSpace(agentfileLayer) == "" {
		return "CONFIG mcp_enabled = true\n"
	}
	return agentfileLayer
}

func appendKnowledgeLayerForTest(
	layer string,
	mounts []KnowledgeMountRequest,
) string {
	out := layer
	for _, mount := range mounts {
		line := "KNOWLEDGE " + mount.Slug
		if strings.TrimSpace(mount.Mode) != "" {
			line += " [" + mount.Mode + "]"
		}
		out = appendPlanLayerLineForTest(out, line)
	}
	return out
}

func appendPlanLayerLineForTest(layer string, line string) string {
	if strings.TrimSpace(layer) == "" {
		return line
	}
	return strings.TrimRight(layer, "\n") + "\n" + line
}

func planSourceModelResourceForTest(
	orchestrator *PodOrchestrator,
	agentSlug string,
) *resourcesvc.ResolvedResource {
	if resolver, ok := orchestrator.modelResources.(*recordingModelResourceResolver); ok &&
		resolver.resource != nil {
		return resolver.resource
	}
	if agentSlug == "claude-code" {
		return resolvedResource("anthropic", "https://api.anthropic.com", "claude-test")
	}
	return resolvedOpenAIResource()
}

func planSourceRepositoryForTest(
	orchestrator *PodOrchestrator,
	req *OrchestrateCreatePodRequest,
) (*gitprovider.Repository, error) {
	if req.RepositoryID == nil {
		return nil, nil
	}
	if selector, ok := orchestrator.runnerSelector.(*mockRunnerSelector); ok &&
		req.RunnerID != 0 &&
		selector.resolveErr != nil {
		return nil, nil
	}
	switch service := orchestrator.repoService.(type) {
	case *mockRepoService:
		return planSourceMockRepositoryForTest(service, req)
	case *routingRepositoryService:
		return planSourceRoutingRepositoryForTest(service, req)
	default:
		return nil, ErrCreateResourceUnavailable
	}
}

func planSourceMockRepositoryForTest(
	service *mockRepoService,
	req *OrchestrateCreatePodRequest,
) (*gitprovider.Repository, error) {
	if service == nil {
		return nil, ErrCreateResourceUnavailable
	}
	if service.err != nil {
		return nil, createRepositoryError(service.err)
	}
	if service.repo == nil {
		return nil, ErrCreateResourceUnavailable
	}
	repo := *service.repo
	repo.ID = *req.RepositoryID
	repo.OrganizationID = req.OrganizationID
	repo.IsActive = true
	return &repo, nil
}

func planSourceRoutingRepositoryForTest(
	service *routingRepositoryService,
	req *OrchestrateCreatePodRequest,
) (*gitprovider.Repository, error) {
	if service == nil || service.byID == nil {
		return nil, ErrCreateResourceUnavailable
	}
	repo := service.byID[*req.RepositoryID]
	if repo == nil {
		return nil, ErrCreateResourceUnavailable
	}
	service.getCalls = append(service.getCalls, *req.RepositoryID)
	clone := *repo
	clone.OrganizationID = req.OrganizationID
	clone.IsActive = true
	return &clone, nil
}

func clearLegacyPlanFixtureFields(req *OrchestrateCreatePodRequest) {
	req.AgentSlug = ""
	req.RepositoryID = nil
	req.Alias = nil
	req.AgentfileLayer = nil
	req.AutomationLevel = ""
	req.BranchName = nil
	req.ModelResourceID = nil
	req.TokenBudget = nil
	req.Perpetual = false
	req.LocalPath = ""
	req.KnowledgeMounts = nil
	req.ModelResourceEnv = nil
	req.ModelResourceArgs = nil
}
