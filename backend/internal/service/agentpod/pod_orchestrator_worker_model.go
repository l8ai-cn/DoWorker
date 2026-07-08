package agentpod

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	"github.com/anthropics/agentsmesh/backend/internal/domain/aimodel"
	aimodelsvc "github.com/anthropics/agentsmesh/backend/internal/service/aimodel"
)

const workerModelBundleName = "worker-model"

// AIModelPoolForOrchestrator resolves a real ai_models row into decrypted
// provider credentials for pod injection.
type AIModelPoolForOrchestrator interface {
	Resolve(ctx context.Context, id int64) (*aimodelsvc.ResolvedModel, error)
}

// VirtualKeyPoolForOrchestrator resolves a virtual API key into the underlying
// ai_models credential plus the key's informational token budget.
type VirtualKeyPoolForOrchestrator interface {
	ResolveModel(ctx context.Context, keyID int64) (*aimodelsvc.ResolvedModel, *int64, error)
}

// applyWorkerModel injects the selected model's credentials into the pod's
// AgentFile layer (config bundle for do-agent, env bundle for codex/claude)
// and appends an informational token_budget CONFIG. Usage is attributed to
// req.VirtualAPIKeyID when a virtual key was selected.
func (o *PodOrchestrator) applyWorkerModel(ctx context.Context, req *OrchestrateCreatePodRequest, agentDef *agentDomain.Agent) error {
	if req.VirtualAPIKeyID == nil && req.ModelConfigID == nil {
		return nil
	}

	resolved, budget, err := o.resolvePoolModel(ctx, req)
	if err != nil {
		return err
	}
	if resolved == nil || resolved.Model == nil {
		return nil
	}

	isDoAgent := agentDef != nil && agentDef.Executable == "do-agent"
	kind := aimodel.HarnessMountKindFor(req.AgentSlug, isDoAgent)
	if kind == aimodel.HarnessMountEnv {
		env := aimodel.HarnessEnvVars(req.AgentSlug, "", resolved.Model, resolved.Credentials)
		if len(env) > 0 {
			if req.SessionEnvBundles == nil {
				req.SessionEnvBundles = map[string]map[string]string{}
			}
			req.SessionEnvBundles[workerModelBundleName] = env
			appendAgentfileLayer(&req.AgentfileLayer, fmt.Sprintf(`USE_ENV_BUNDLE "%s"`, workerModelBundleName))
		}
	} else {
		if req.SessionConfigBundles == nil {
			req.SessionConfigBundles = map[string]interface{}{}
		}
		req.SessionConfigBundles[workerModelBundleName] = resolved.SettingsJSON("")
		appendAgentfileLayer(&req.AgentfileLayer, fmt.Sprintf(`USE_CONFIG_BUNDLE "%s"`, workerModelBundleName))
	}

	tb := req.TokenBudget
	if tb == nil {
		tb = budget
	}
	if tb != nil && *tb > 0 {
		appendAgentfileLayer(&req.AgentfileLayer, fmt.Sprintf(`CONFIG token_budget = "%s"`, strconv.FormatInt(*tb, 10)))
	}
	return nil
}

func (o *PodOrchestrator) resolvePoolModel(ctx context.Context, req *OrchestrateCreatePodRequest) (*aimodelsvc.ResolvedModel, *int64, error) {
	if req.VirtualAPIKeyID != nil {
		if o.virtualKeyPool == nil {
			return nil, nil, nil
		}
		return o.virtualKeyPool.ResolveModel(ctx, *req.VirtualAPIKeyID)
	}
	if o.aiModelPool == nil {
		return nil, nil, nil
	}
	resolved, err := o.aiModelPool.Resolve(ctx, *req.ModelConfigID)
	return resolved, nil, err
}

func appendAgentfileLayer(layer **string, lines ...string) {
	var base string
	if *layer != nil {
		base = **layer
	}
	parts := make([]string, 0, len(lines)+1)
	if strings.TrimSpace(base) != "" {
		parts = append(parts, base)
	}
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			parts = append(parts, l)
		}
	}
	out := strings.Join(parts, "\n")
	*layer = &out
}
