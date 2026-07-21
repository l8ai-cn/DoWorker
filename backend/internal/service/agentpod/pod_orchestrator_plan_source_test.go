package agentpod

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/l8ai-cn/agentcloud/agentfile"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/gitprovider"
	resourcesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
	workercreation "github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
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
	artifactLayer := planSourceArtifactLayer(spec, agentfileLayer)
	artifact, dependencies := planArtifactForTest(
		t,
		ctx,
		&spec,
		agentfileLayer,
		resource,
		nil,
	)
	draft := workercreation.Draft{
		OptionsRevision: "test-options",
		WorkerSpec:      workerSpecDraftFromSpec(spec),
	}
	prepared := workercreation.Prepared{
		Snapshot:       resolvedWorkerSpecFromSpecForPodServiceTest(t, ctxOrgID(ctx), spec),
		Spec:           spec,
		AgentfileLayer: artifactLayer,
		Artifact:       artifact,
		Dependencies:   dependencies,
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
	repositoryDeclaration, err := parseRepositoryDeclaration(layer)
	if err != nil {
		materializeFailedPlanSourceForTest(
			req,
			orchestrator,
			fmt.Errorf("%w: %v", ErrInvalidAgentfileLayer, err),
		)
		return
	}
	if repositoryDeclaration.slug != "" && req.RepositoryID == nil {
		materializeFailedPlanSourceForTest(
			req,
			orchestrator,
			ErrCreateResourceUnavailable,
		)
		return
	}
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
		materializeFailedPlanSourceForTest(req, orchestrator, prepareErr)
		return
	}
	if prepareErr == nil && repository == nil && req.RepositoryID != nil {
		spec.Workspace.RepositoryID = nil
		spec.Workspace.Branch = ""
	}
	if repository != nil && strings.TrimSpace(spec.Workspace.Branch) == "" {
		spec.Workspace.Branch = firstNonEmpty(repository.DefaultBranch, "main")
	}
	artifact, dependencies := planArtifactForTest(
		t,
		context.WithValue(
			context.WithValue(context.Background(), ctxKeyOrgID, req.OrganizationID),
			ctxKeyUserID,
			req.UserID,
		),
		&spec,
		layer,
		resource,
		repository,
		orchestrator.userService,
	)
	req.WorkerSpecDraft = &workercreation.Draft{
		OptionsRevision: "test-options",
		WorkerSpec:      workerSpecDraftFromSpec(spec),
	}
	prepared := workercreation.Prepared{
		Snapshot:       resolvedWorkerSpecFromSpecForPodServiceTest(t, req.OrganizationID, spec),
		Spec:           spec,
		AgentfileLayer: planSourceArtifactLayer(spec, layer),
		Repository:     repository,
		Artifact:       artifact,
		Dependencies:   dependencies,
	}
	orchestrator.workerCreation = &workerCreationPreparer{
		prepared: prepared,
		err:      prepareErr,
	}
	clearLegacyPlanFixtureFields(req)
}

func materializeFailedPlanSourceForTest(
	req *OrchestrateCreatePodRequest,
	orchestrator *PodOrchestrator,
	err error,
) {
	spec := workerSpecForPlanSource(
		req.AgentSlug,
		workerSpecStringValue(req.AgentfileLayer),
		planSourceModelResourceForTest(orchestrator, req.AgentSlug),
	)
	req.WorkerSpecDraft = &workercreation.Draft{
		OptionsRevision: "test-options",
		WorkerSpec:      workerSpecDraftFromSpec(spec),
	}
	orchestrator.workerCreation = &workerCreationPreparer{err: err}
	clearLegacyPlanFixtureFields(req)
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
