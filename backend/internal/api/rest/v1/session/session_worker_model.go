package sessionapi

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	aimodeldomain "github.com/anthropics/agentsmesh/backend/internal/domain/aimodel"
	aimodelsvc "github.com/anthropics/agentsmesh/backend/internal/service/aimodel"
	"github.com/gin-gonic/gin"
)

// isWorkerAgent reports whether the agent slug launches a do-agent Worker.
func (d *Deps) isWorkerAgent(ctx context.Context, slug string) bool {
	if slug == "do-agent" {
		return true
	}
	if d.Agent == nil {
		return false
	}
	a, err := d.Agent.GetBySlug(ctx, slug)
	if err != nil || a == nil {
		return false
	}
	return a.Executable == "do-agent"
}

// resolveWorkerModel mounts a pool model into the new Worker: do-agent gets
// settings.json (USE_CONFIG_BUNDLE); codex-cli gets credential env
// (USE_ENV_BUNDLE). Selection: model_config_id, then harness default, then org default.
func (d *Deps) resolveWorkerModel(
	c *gin.Context,
	userID, orgID int64,
	body createSessionBody,
	layer **string,
) (*workerModelMount, error) {
	if d.AIModels == nil {
		return nil, nil
	}
	ctx := c.Request.Context()
	isDoAgent := d.isWorkerAgent(ctx, body.AgentID)
	kind := aimodeldomain.HarnessMountKindFor(body.AgentID, isDoAgent)
	if body.ModelConfigID == nil && kind == aimodeldomain.HarnessMountNone {
		return nil, nil
	}

	var resolved *aimodelsvc.ResolvedModel
	var err error
	if body.ModelConfigID != nil {
		resolved, err = d.AIModels.ResolveVisible(ctx, *body.ModelConfigID, userID, orgID)
		if err != nil {
			if err == aimodelsvc.ErrNotFound {
				return nil, fmt.Errorf("selected model not found")
			}
			return nil, err
		}
	} else {
		resolved, err = d.AIModels.ResolveDefaultForAgent(ctx, userID, orgID, body.AgentID)
		if err != nil {
			return nil, err
		}
		if resolved == nil {
			return nil, nil
		}
	}

	overrideModel := ""
	if body.Model != nil {
		overrideModel = strings.TrimSpace(*body.Model)
	}

	mount := &workerModelMount{}
	switch kind {
	case aimodeldomain.HarnessMountConfig:
		mount.ConfigBundles = map[string]interface{}{
			workerModelBundleName: resolved.SettingsJSON(overrideModel),
		}
		appendLayerLines(layer, fmt.Sprintf(`USE_CONFIG_BUNDLE "%s"`, workerModelBundleName))
	case aimodeldomain.HarnessMountEnv:
		env := aimodeldomain.HarnessEnvVars(body.AgentID, overrideModel, resolved.Model, resolved.Credentials)
		if len(env) == 0 {
			return nil, nil
		}
		mount.EnvBundles = map[string]map[string]string{workerModelBundleName: env}
		appendLayerLines(layer, fmt.Sprintf(`USE_ENV_BUNDLE "%s"`, workerModelBundleName))
	default:
		if body.ModelConfigID == nil {
			return nil, nil
		}
		mount.ConfigBundles = map[string]interface{}{
			workerModelBundleName: resolved.SettingsJSON(overrideModel),
		}
		appendLayerLines(layer, fmt.Sprintf(`USE_CONFIG_BUNDLE "%s"`, workerModelBundleName))
	}

	if body.TokenBudget != nil && *body.TokenBudget > 0 {
		appendLayerLines(layer, fmt.Sprintf(`CONFIG token_budget = "%s"`, strconv.FormatInt(*body.TokenBudget, 10)))
	}
	return mount, nil
}

func appendLayerLines(layer **string, lines ...string) {
	var base string
	if *layer != nil {
		base = **layer
	}
	parts := []string{}
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
