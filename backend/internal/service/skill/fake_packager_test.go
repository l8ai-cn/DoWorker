package skill

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
)

type fakePackager struct {
	lastSkillMd  string
	lastSkillCfg string
	calls        int
	failErr      error
	reused       bool
	deletedKeys  []string
	deleteErr    error
	deleteHook   func()
}

func (p *fakePackager) PackageFromDir(_ context.Context, dir string) (*extensionsvc.PackagedSkill, error) {
	if p.failErr != nil {
		return nil, p.failErr
	}
	p.calls++
	md, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		return nil, fmt.Errorf("SKILL.md missing in materialized dir: %w", err)
	}
	cfg, err := os.ReadFile(filepath.Join(dir, "skill.json"))
	if err != nil {
		return nil, fmt.Errorf("skill.json missing in materialized dir: %w", err)
	}
	p.lastSkillMd = string(md)
	p.lastSkillCfg = string(cfg)

	sum := sha256.Sum256(append(md, cfg...))
	sha := fmt.Sprintf("%x", sum)
	slug := frontmatterName(string(md))
	return &extensionsvc.PackagedSkill{
		Slug:        slug,
		DisplayName: slug,
		ContentSha:  sha,
		StorageKey:  fmt.Sprintf("skills/direct/%s/%s.tar.gz", slug, sha),
		PackageSize: int64(len(md) + len(cfg)),
		Created:     !p.reused,
	}, nil
}

func frontmatterName(md string) string {
	for _, line := range strings.Split(md, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		}
	}
	return ""
}
