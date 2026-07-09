package extension

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SkillSourceAuth carries optional credentials for cloning private skill
// sources. Type is one of the AuthType* constants.
type SkillSourceAuth struct {
	Type       string
	Credential string
}

// ClonedSkillSource is a shallow clone of an external skill repo on local
// disk. Callers MUST Close() it to release the temp directory.
type ClonedSkillSource struct {
	Dir       string
	CommitSha string
	cleanup   func()
}

func (c *ClonedSkillSource) Close() {
	if c.cleanup != nil {
		c.cleanup()
	}
}

// CloneSkillSource shallow-clones url@branch into a temp dir. This is the
// single entry point the skill service uses to pull external skill repos
// (mainstream skill hubs are git repos of SKILL.md directories).
func CloneSkillSource(ctx context.Context, url, branch string, auth *SkillSourceAuth) (*ClonedSkillSource, error) {
	tmpDir, err := os.MkdirTemp("", "skill-import-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }

	repoDir := filepath.Join(tmpDir, "repo")
	if auth != nil && auth.Type != "" && auth.Type != AuthTypeNone && auth.Credential != "" {
		err = gitCloneWithAuth(ctx, url, branch, repoDir, auth.Type, auth.Credential)
	} else {
		err = gitClone(ctx, url, branch, repoDir)
	}
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	sha, err := gitHead(ctx, repoDir)
	if err != nil {
		sha = ""
	}
	return &ClonedSkillSource{Dir: repoDir, CommitSha: sha, cleanup: cleanup}, nil
}

// ScanSkillSource discovers skills in a cloned source. When subdir is set,
// only that directory is parsed (it must contain a SKILL.md). Otherwise the
// layout is auto-detected: a root SKILL.md is a single-skill repo; anything
// else is scanned as a collection (skills/ subdirs or sibling dirs).
// The returned SkillInfo.DirPath values point inside the clone.
func ScanSkillSource(repoDir, subdir string) ([]SkillInfo, error) {
	if subdir != "" {
		dir := filepath.Join(repoDir, filepath.Clean(subdir))
		if !strings.HasPrefix(dir, filepath.Clean(repoDir)+string(os.PathSeparator)) {
			return nil, fmt.Errorf("invalid subdir: escapes repository directory")
		}
		if !fileExists(filepath.Join(dir, "SKILL.md")) {
			return nil, fmt.Errorf("SKILL.md not found in %q", subdir)
		}
		info, err := parseSkillDir(dir)
		if err != nil {
			return nil, err
		}
		return []SkillInfo{*info}, nil
	}

	if detectRepoType(repoDir) == "single" {
		info, err := parseSkillDir(repoDir)
		if err != nil {
			return nil, err
		}
		return []SkillInfo{*info}, nil
	}
	return scanCollectionSkills(repoDir)
}

// SkillSourceSubdir returns the skill directory's path relative to the
// clone root ("" when the skill is the repo root), suitable for persisting
// as upstream_subdir.
func SkillSourceSubdir(repoDir string, info SkillInfo) string {
	rel, err := filepath.Rel(repoDir, info.DirPath)
	if err != nil || rel == "." {
		return ""
	}
	return filepath.ToSlash(rel)
}
