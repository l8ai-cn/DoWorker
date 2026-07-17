package expert

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/lib/pq"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

func validateExpertMatchesWorkerSpec(
	source *expertdom.Expert,
	spec specdomain.Spec,
	skills []skilldom.Skill,
) error {
	config, err := decodeMarketConfig(source.ConfigOverrides)
	if err != nil {
		return err
	}
	mounts, err := decodeMarketKnowledgeMounts(source.KnowledgeMounts)
	if err != nil {
		return err
	}
	sourceSkills := sortedStrings(source.SkillSlugs)
	specSkills := marketSkillSlugs(skills)
	switch {
	case source.AgentSlug != spec.Runtime.WorkerType.Slug.String():
		return snapshotDrift("agent_slug")
	case optionalStringValue(source.Prompt) != spec.Workspace.Instructions:
		return snapshotDrift("prompt")
	case source.InteractionMode != string(spec.TypeConfig.InteractionMode):
		return snapshotDrift("interaction_mode")
	case source.AutomationLevel != string(spec.TypeConfig.AutomationLevel):
		return snapshotDrift("automation_level")
	case !reflect.DeepEqual(config, normalizedMarketConfig(spec.TypeConfig.Values)):
		return snapshotDrift("config_overrides")
	case !reflect.DeepEqual(sourceSkills, specSkills):
		return snapshotDrift("skill_slugs")
	case source.Perpetual:
		return snapshotDrift("perpetual")
	case source.RunnerID != nil || source.RepositoryID != nil ||
		source.BranchName != nil || len(source.UsedEnvBundles) > 0 ||
		len(mounts) > 0 || source.AgentfileLayer != nil:
		return fmt.Errorf("expert contains non-portable runtime fields")
	default:
		return nil
	}
}

func snapshotDrift(field string) error {
	return fmt.Errorf("%s does not match workerspec snapshot", field)
}

func optionalMarketPrompt(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func sortedStrings(values pq.StringArray) []string {
	sorted := append([]string{}, values...)
	sort.Strings(sorted)
	return sorted
}

func normalizedMarketConfig(values map[string]any) map[string]any {
	if values == nil {
		return map[string]any{}
	}
	return values
}
