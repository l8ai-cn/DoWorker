package skill

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	skilldom "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
	extensionsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/extension"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/gitops"
)

// jsonMarshal indirection keeps importing.go free of encoding/json noise.
func jsonMarshal(v any) ([]byte, error) { return json.Marshal(v) }

// ErrNotImported is returned when SyncFromUpstream targets a skill without
// upstream provenance (platform-authored rows have nothing to sync from).
var ErrNotImported = errors.New("skill: not imported from an upstream repo")

// SyncFromUpstream re-clones the recorded upstream (default branch), locates
// the skill's subdir, and refreshes the internal repo + package + catalog row.
func (s *Service) SyncFromUpstream(ctx context.Context, orgID int64, slug string) (*skilldom.Skill, error) {
	row, err := s.store.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if row.UpstreamURL == "" {
		return nil, ErrNotImported
	}
	if _, err := ValidateTags(row.Tags); err != nil {
		return nil, err
	}

	src, err := extensionsvc.CloneSkillSource(ctx, row.UpstreamURL, "", nil)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	infos, err := extensionsvc.ScanSkillSource(src.Dir, row.UpstreamSubdir)
	if err != nil {
		return nil, fmt.Errorf("skill: upstream no longer contains %q: %w", row.UpstreamSubdir, err)
	}
	info := infos[0]

	files, err := readSkillDirFiles(info.DirPath)
	if err != nil {
		return nil, err
	}
	return s.refreshImportedSkill(ctx, row, src, info, files)
}

func prepareImportedSkillFiles(
	slug string,
	tags []string,
	files []gitops.FileChange,
) ([]gitops.FileChange, error) {
	normalized, err := ValidateTags(tags)
	if err != nil {
		return nil, err
	}
	synced, err := preserveCuratorTags(files, slug, normalized)
	if err != nil {
		return nil, err
	}
	for _, file := range synced {
		if file.Path == "skill.json" {
			return synced, nil
		}
	}
	return nil, errors.New("skill: synchronized skill.json is missing")
}

func preserveCuratorTags(files []gitops.FileChange, slug string, tags []string) ([]gitops.FileChange, error) {
	normalized := []string(skilldom.NormalizeTags(tags))
	synced := append([]gitops.FileChange(nil), files...)
	for i := range synced {
		if synced[i].Path != "skill.json" {
			continue
		}
		content, err := replaceSkillConfigFields(synced[i].Content, map[string]any{
			"schema": skillConfigSchema,
			"slug":   slug,
			"tags":   normalized,
		})
		if err != nil {
			return nil, fmt.Errorf("skill: synchronize skill.json: %w", err)
		}
		synced[i].Content = content
		return synced, nil
	}

	content, err := json.MarshalIndent(skillConfig{
		Schema: skillConfigSchema,
		Slug:   slug,
		Tags:   normalized,
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("skill: render synchronized skill.json: %w", err)
	}
	return append(synced, gitops.FileChange{
		Path:    "skill.json",
		Content: append(content, '\n'),
	}), nil
}
