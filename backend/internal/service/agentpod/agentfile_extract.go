package agentpod

import (
	"fmt"

	"github.com/l8ai-cn/agentcloud/agentfile"
	"github.com/l8ai-cn/agentcloud/agentfile/extract"
	"github.com/l8ai-cn/agentcloud/agentfile/merge"
	"github.com/l8ai-cn/agentcloud/agentfile/parser"
	"github.com/l8ai-cn/agentcloud/agentfile/resolve"
	"github.com/l8ai-cn/agentcloud/agentfile/serialize"
	agentDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agent"
)

type agentfileExtractResult struct {
	Mode                  string                    // MODE pty/acp
	Branch                string                    // BRANCH "branch-name"
	RepoSlug              string                    // REPO "slug" (e.g., "dev-org/demo-api")
	Prompt                string                    // PROMPT "prompt content"
	Knowledge             []agentfile.KnowledgeSpec // KNOWLEDGE slug [mode], ...
	ConfigValues          agentDomain.ConfigValues
	MergedAgentfileSource string
}

func extractFromAgentfileLayer(
	baseAgentfileSrc, userLayerSrc string,
	userPrefs, systemOverrides map[string]interface{},
) (*agentfileExtractResult, error) {
	baseProg, baseErrs := parser.Parse(baseAgentfileSrc)
	if len(baseErrs) > 0 {
		return nil, fmt.Errorf("base agentfile parse error: %v", baseErrs[0])
	}

	userProg, userErrs := parser.Parse(userLayerSrc)
	if len(userErrs) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAgentfileLayer, userErrs[0])
	}

	layerConfigNames := resolve.ExtractConfigNames(userProg)

	merge.Merge(baseProg, userProg)

	resolve.ResolveConfigValues(baseProg, layerConfigNames, userPrefs, systemOverrides)

	mergedSource := serialize.Serialize(baseProg)
	spec := extract.Extract(baseProg)

	result := &agentfileExtractResult{
		Mode:                  spec.Mode,
		Prompt:                spec.Prompt,
		Knowledge:             spec.Knowledge,
		MergedAgentfileSource: mergedSource,
		ConfigValues:          make(agentDomain.ConfigValues),
	}

	if spec.Repo != nil {
		result.RepoSlug = spec.Repo.URL
		result.Branch = spec.Repo.Branch
	}

	for _, cfg := range spec.Config {
		if !isSystemConfigKey(cfg.Name) {
			result.ConfigValues[cfg.Name] = cfg.Default
		}
	}

	return result, nil
}
