package agentpod

import (
	"strings"

	"github.com/anthropics/agentsmesh/agentfile/parser"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	resourcesvc "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

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
	if strings.Contains(agentfileLayer, `USE_ENV_BUNDLE "runtime"`) {
		spec.Workspace.EnvBundleIDs = []specdomain.RuntimeEnvBundleID{1}
	}
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
	if mounts := planKnowledgeMountsForTest(
		workerSpecStringValue(req.AgentfileLayer),
		req.KnowledgeMounts,
	); len(mounts) > 0 {
		spec.Workspace.KnowledgeMounts = mounts
	}
}

func planKnowledgeMountsForTest(
	layer string,
	requests []KnowledgeMountRequest,
) []specdomain.KnowledgeMount {
	mounts := []specdomain.KnowledgeMount{}
	for _, line := range strings.Split(layer, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "KNOWLEDGE" {
			continue
		}
		mode := "ro"
		if len(fields) >= 3 {
			mode = strings.Trim(fields[2], "[]")
		}
		mounts = append(mounts, planKnowledgeMountForSlug(fields[1], mode))
	}
	for _, request := range requests {
		mode := strings.TrimSpace(request.Mode)
		if mode == "" {
			mode = "ro"
		}
		mounts = append(mounts, planKnowledgeMountForSlug(request.Slug, mode))
	}
	return mounts
}

func planKnowledgeMountForSlug(slug string, mode string) specdomain.KnowledgeMount {
	return specdomain.KnowledgeMount{
		KnowledgeBaseID: planKnowledgeIDForSlug(slug),
		Mode:            specdomain.KnowledgeMountMode(mode),
	}
}

func planKnowledgeIDForSlug(slug string) int64 {
	switch slug {
	case "team-docs":
		return 101
	case "product-wiki":
		return 102
	default:
		return 1000 + int64(len(slug))
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
	_ string,
	agentfileLayer string,
	_ *PodOrchestrator,
) specdomain.InteractionMode {
	program, parseErrors := parser.Parse(agentfileLayer)
	if len(parseErrors) == 0 {
		for _, declaration := range program.Declarations {
			if mode, ok := declaration.(*parser.ModeDecl); ok {
				return specdomain.InteractionMode(mode.Mode)
			}
		}
	}
	return specdomain.InteractionModeACP
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
		if agentSlug != "claude-code" &&
			resolver.resource.Connection.ProviderKey.String() == "anthropic" {
			return resolvedOpenAIResource()
		}
		return resolver.resource
	}
	if agentSlug == "claude-code" {
		return resolvedResource("anthropic", "https://api.anthropic.com", "claude-test")
	}
	return resolvedOpenAIResource()
}
