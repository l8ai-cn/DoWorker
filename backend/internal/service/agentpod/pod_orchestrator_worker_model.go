package agentpod

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/anthropics/agentsmesh/agentfile"
	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	resourceDomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	resourcesvc "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
)

const workerModelBundleName = "worker-model"

type ModelResourceResolver interface {
	ResolveExact(context.Context, resourcesvc.Actor, int64, int64, resourcesvc.ResolutionRequirements) (*resourcesvc.ResolvedResource, error)
}

func (o *PodOrchestrator) applyWorkerModel(ctx context.Context, req *OrchestrateCreatePodRequest, agentDef *agentDomain.Agent) error {
	harness := workerModelHarness(req.AgentSlug, agentDef)
	requirements, needsResource := modelRequirementsForRequest(req, agentDef)
	if !needsResource {
		return nil
	}
	if req.ModelResourceID == nil || *req.ModelResourceID <= 0 {
		return ErrMissingModelResource
	}
	if o.modelResources == nil {
		return ErrModelResourceResolverUnavailable
	}
	resource, err := o.modelResources.ResolveExact(
		ctx,
		resourcesvc.Actor{UserID: req.UserID},
		req.OrganizationID,
		*req.ModelResourceID,
		requirements,
	)
	if err != nil {
		return err
	}
	if req.preparedWorkerSpec != nil {
		if err := validatePreparedModelBinding(
			req.preparedWorkerSpec.Runtime.ModelBinding,
			resource,
		); err != nil {
			return err
		}
	}
	if harness == "do-agent" {
		settings, err := doAgentModelSettings(resource)
		if err != nil {
			return err
		}
		if req.SessionConfigBundles == nil {
			req.SessionConfigBundles = map[string]interface{}{}
		}
		req.SessionConfigBundles[workerModelBundleName] = settings
		appendAgentfileLayer(&req.AgentfileLayer, fmt.Sprintf(`USE_CONFIG_BUNDLE "%s"`, workerModelBundleName))
	} else {
		env, err := modelResourceEnvironment(harness, resource)
		if err != nil {
			return err
		}
		req.ModelResourceEnv = env
		modelID := strings.TrimSpace(resource.Resource.ModelID)
		if harness == "claude-code" && modelID != "" {
			appendAgentfileLayer(&req.AgentfileLayer, `CONFIG model = `+agentfile.FormatStringLiteral(resource.Resource.ModelID))
		}
		if harness == "openclaw" && modelID != "" {
			appendAgentfileLayer(&req.AgentfileLayer, `CONFIG model = `+agentfile.FormatStringLiteral(resource.Resource.ModelID))
		}
		if harness == "gemini-cli" {
			if modelID == "" {
				return ErrMissingModelResource
			}
			req.ModelResourceArgs = []string{"--model", modelID}
		}
	}
	if req.TokenBudget != nil && *req.TokenBudget > 0 {
		appendAgentfileLayer(&req.AgentfileLayer, fmt.Sprintf(`CONFIG token_budget = "%s"`, strconv.FormatInt(*req.TokenBudget, 10)))
	}
	return nil
}

func modelRequirementsForRequest(
	req *OrchestrateCreatePodRequest,
	agentDef *agentDomain.Agent,
) (resourcesvc.ResolutionRequirements, bool) {
	if req != nil && req.preparedWorkerSpec != nil {
		binding := req.preparedWorkerSpec.Runtime.ModelBinding
		if binding.IsEmpty() {
			return resourcesvc.ResolutionRequirements{}, false
		}
		return chatRequirements(binding.ProtocolAdapter.String()), true
	}
	return modelResourceRequirements(req.AgentSlug, agentDef)
}

func validatePreparedModelBinding(
	expected specdomain.ModelBinding,
	resolved *resourcesvc.ResolvedResource,
) error {
	if resolved == nil ||
		resolved.Resource.ID != expected.ResourceID ||
		resolved.Resource.Revision != expected.ResourceRevision ||
		resolved.Connection.ID != expected.ConnectionID ||
		resolved.Connection.Revision != expected.ConnectionRevision ||
		resolved.Connection.ProviderKey != expected.ProviderKey ||
		resolved.Provider.ProtocolAdapter != expected.ProtocolAdapter.String() ||
		strings.TrimSpace(resolved.Resource.ModelID) != expected.ModelID {
		return ErrWorkerSpecModelChanged
	}
	return nil
}

func modelResourceRequirements(agentSlug string, agentDef *agentDomain.Agent) (resourcesvc.ResolutionRequirements, bool) {
	switch workerModelHarness(agentSlug, agentDef) {
	case "do-agent":
		return chatRequirements("openai-compatible", "anthropic", "minimax"), true
	case "codex-cli":
		return chatRequirements("openai-compatible"), true
	case "claude-code":
		return chatRequirements("anthropic"), true
	case "gemini-cli":
		return chatRequirements("gemini"), true
	case "openclaw", "hermes":
		return chatRequirements("openai-compatible", "anthropic", "gemini"), true
	default:
		return resourcesvc.ResolutionRequirements{}, false
	}
}

func workerModelHarness(agentSlug string, agentDef *agentDomain.Agent) string {
	value := agentSlug
	if agentDef != nil && strings.TrimSpace(agentDef.Executable) != "" {
		value = agentDef.Executable
	}
	switch strings.TrimSpace(value) {
	case "codex", "codex-cli":
		return "codex-cli"
	case "claude", "claude-code":
		return "claude-code"
	case "gemini", "gemini-cli":
		return "gemini-cli"
	case "do-agent":
		return "do-agent"
	case "openclaw":
		return "openclaw"
	case "hermes":
		return "hermes"
	default:
		return ""
	}
}

func chatRequirements(adapters ...string) resourcesvc.ResolutionRequirements {
	return resourcesvc.ResolutionRequirements{
		Modality:                resourceDomain.ModalityChat,
		Capability:              resourceDomain.CapabilityTextGeneration,
		AllowedProtocolAdapters: adapters,
	}
}

func appendAgentfileLayer(layer **string, lines ...string) {
	parts := make([]string, 0, len(lines)+1)
	if layer != nil && *layer != nil && strings.TrimSpace(**layer) != "" {
		parts = append(parts, **layer)
	}
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			parts = append(parts, line)
		}
	}
	if len(parts) == 0 {
		return
	}
	result := strings.Join(parts, "\n")
	*layer = &result
}
