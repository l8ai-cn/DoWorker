package skill

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
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
