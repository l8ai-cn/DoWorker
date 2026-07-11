package agent

import (
	"context"
	"fmt"

	"github.com/anthropics/agentsmesh/agentfile/eval"
	"github.com/anthropics/agentsmesh/agentfile/parser"
	"github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

// Sandbox path placeholders — Runner replaces with real paths after sandbox setup.
const (
	PlaceholderSandboxRoot = "{{sandbox_root}}"
	PlaceholderWorkDir     = "{{work_dir}}"
)

// buildFromAgentfile evaluates the agent's AgentFile with placeholder sandbox
// paths and produces a CreatePodCommand. Credential injection is handled by
// AgentFile USE_ENV_BUNDLE declarations referencing entries in the eval
// context's EnvBundles map; the backend no longer threads a parallel
// credential blob through CreatePodCommand.
func (b *ConfigBuilder) buildFromAgentfile(
	ctx context.Context,
	req *ConfigBuildRequest,
	agentDef *agent.Agent,
) (*runnerv1.CreatePodCommand, error) {
	mergedSource := req.MergedAgentfileSource
	if mergedSource == "" {
		return nil, fmt.Errorf("agent %s: MergedAgentfileSource is empty (AgentFile resolve should always produce it)", req.AgentSlug)
	}

	// Build MCP context
	builtinMCP, installedMCP := b.buildMCPContext(ctx, req, agentDef.Slug)

	// Build EnvBundle context (mirror of MCP pattern: load every visible
	// bundle, decrypt, expose by name to eval; USE_ENV_BUNDLE picks).
	envBundles, err := b.buildEnvBundleContext(ctx, req, agentDef.Slug)
	if err != nil {
		return nil, err
	}
	configBundles := b.buildConfigBundleContext(ctx, req, agentDef.Slug)

	// Parse and eval AgentFile with placeholder context
	prog, errs := parser.Parse(mergedSource)
	if len(errs) > 0 {
		return nil, fmt.Errorf("agentfile parse error: %v", errs[0])
	}

	evalCtx := buildEvalContext(req, builtinMCP, installedMCP, envBundles, configBundles)
	if err := eval.Eval(prog, evalCtx); err != nil {
		return nil, fmt.Errorf("agentfile eval error: %w", err)
	}
	eval.ApplyModeArgs(evalCtx.Result)
	eval.ApplyRemoves(evalCtx.Result)

	// AgentFile SETUP is the most specific source for preparation scripts.
	// Preserve repository-level preparation as a fallback when SETUP is absent.
	effectiveReq := *req
	if evalCtx.Result.Setup.Script != "" {
		effectiveReq.PreparationScript = evalCtx.Result.Setup.Script
		effectiveReq.PreparationTimeout = evalCtx.Result.Setup.Timeout
	}

	cmd := buildResultToProto(&effectiveReq, evalCtx.Result)
	cmd.ResourcesToDownload, err = b.buildSkillResources(
		ctx,
		req,
		agentDef.Slug,
		evalCtx.Result.Skills,
	)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}
