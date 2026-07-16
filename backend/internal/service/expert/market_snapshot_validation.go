package expert

import (
	"encoding/json"
	"fmt"
	"strings"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func decodeMarketConfig(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var config map[string]any
	if err := decodeStrictJSON(raw, &config); err != nil {
		return nil, err
	}
	if config == nil {
		return nil, fmt.Errorf("config_overrides must be an object")
	}
	return config, nil
}

func decodeMarketKnowledgeMounts(
	raw json.RawMessage,
) ([]expertdom.KnowledgeMount, error) {
	if len(raw) == 0 {
		return []expertdom.KnowledgeMount{}, nil
	}
	var mounts []expertdom.KnowledgeMount
	if err := decodeStrictJSON(raw, &mounts); err != nil {
		return nil, err
	}
	if mounts == nil {
		return nil, fmt.Errorf("knowledge_mounts must be an array")
	}
	if err := validateMarketKnowledgeMounts(mounts); err != nil {
		return nil, err
	}
	return mounts, nil
}

func validateMarketExpertSnapshot(snapshot marketExpertSnapshot) error {
	if strings.TrimSpace(snapshot.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(snapshot.AgentSlug) == "" {
		return fmt.Errorf("agent_slug is required")
	}
	if err := slugkit.Validate(snapshot.Slug); err != nil {
		return fmt.Errorf("slug: %w", err)
	}
	if err := slugkit.Validate(snapshot.AgentSlug); err != nil {
		return fmt.Errorf("agent_slug: %w", err)
	}
	switch snapshot.InteractionMode {
	case expertdom.InteractionModePTY, expertdom.InteractionModeACP:
	default:
		return fmt.Errorf("invalid interaction_mode %q", snapshot.InteractionMode)
	}
	switch snapshot.AutomationLevel {
	case expertdom.AutomationLevelInteractive,
		expertdom.AutomationLevelAutoEdit,
		expertdom.AutomationLevelAutonomous:
	default:
		return fmt.Errorf("invalid automation_level %q", snapshot.AutomationLevel)
	}
	if snapshot.UsedEnvBundles == nil || snapshot.SkillSlugs == nil ||
		snapshot.KnowledgeMounts == nil || snapshot.ConfigOverrides == nil {
		return fmt.Errorf("snapshot collections must not be null")
	}
	if len(snapshot.UsedEnvBundles) > 0 || len(snapshot.KnowledgeMounts) > 0 {
		return fmt.Errorf("organization-scoped expert references are not portable")
	}
	for _, skillSlug := range snapshot.SkillSlugs {
		if err := slugkit.Validate(skillSlug); err != nil {
			return fmt.Errorf("skill slug %q: %w", skillSlug, err)
		}
	}
	if err := validateMarketKnowledgeMounts(snapshot.KnowledgeMounts); err != nil {
		return err
	}
	return validateMarketMetadata(snapshot.Metadata)
}

func validMarketIcon(icon string) bool {
	switch strings.TrimSpace(icon) {
	case "rocket", "network", "git-compare",
		"clapperboard", "scissors", "film":
		return true
	default:
		return false
	}
}

func validatePortableMarketSpec(spec specdomain.Spec) error {
	switch {
	case spec.Workspace.RepositoryID != nil:
		return fmt.Errorf("repository reference is not portable")
	case len(spec.Workspace.KnowledgeMounts) > 0:
		return fmt.Errorf("knowledge references are not portable")
	case len(spec.Workspace.EnvBundleIDs) > 0:
		return fmt.Errorf("environment references are not portable")
	case len(spec.Workspace.ConfigBundleIDs) > 0:
		return fmt.Errorf("workspace.config_bundle_ids are not portable")
	case len(spec.TypeConfig.SecretRefs) > 0:
		return fmt.Errorf("secret references are not portable")
	default:
		return nil
	}
}

func validateMarketKnowledgeMounts(
	mounts []expertdom.KnowledgeMount,
) error {
	for _, mount := range mounts {
		if err := slugkit.Validate(mount.Slug); err != nil {
			return fmt.Errorf("knowledge mount slug %q: %w", mount.Slug, err)
		}
		switch mount.Mode {
		case "", "ro", "rw":
		default:
			return fmt.Errorf("invalid knowledge mount mode %q", mount.Mode)
		}
	}
	return nil
}

func validateMarketMetadata(raw json.RawMessage) error {
	var metadata map[string]json.RawMessage
	if err := decodeStrictJSON(raw, &metadata); err != nil {
		return err
	}
	if metadata == nil {
		return fmt.Errorf("metadata must be an object")
	}
	return nil
}
