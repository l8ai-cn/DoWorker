package extension

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// PackageFromDir packages an already-materialized skill directory (must contain
// a SKILL.md at its root) via the existing filesystem-based pipeline
// (parseSkillDir + computeDirSHA + packageSkillDir) and uploads the artifact to
// object storage. This is the bridge the git-backed skill service uses after
// materializing an am-skills repo tree into a temp dir — the packager stays
// filesystem-based and unaware of gitops.
//
// It is a thin, exported wrapper over the private packageDir used by
// PackageFromGitHub / PackageFromUpload; those flows are unchanged.
func (p *SkillPackager) PackageFromDir(ctx context.Context, dir string) (*PackagedSkill, error) {
	if !fileExists(filepath.Join(dir, "SKILL.md")) {
		return nil, fmt.Errorf("SKILL.md not found in %s", dir)
	}
	if !dirExists(dir) {
		return nil, fmt.Errorf("skill dir %q not found", dir)
	}
	if _, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("skill dir: %w", err)
	}
	return p.packageDir(ctx, dir)
}
