package skill

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	extensionsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/extension"
)

type fakePackager struct {
	lastSkillMd       string
	lastSkillCfg      string
	calls             int
	catalogIdentities []string
	failErr           error
	reused            bool
	deletedKeys       []string
	deleteErr         error
	deleteHook        func()
}

func (p *fakePackager) PrepareFromDir(
	_ context.Context,
	dir string,
) (*extensionsvc.PreparedSkill, error) {
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
	return &extensionsvc.PreparedSkill{
		Slug:        slug,
		DisplayName: slug,
		ContentSha:  sha,
		StorageKey:  fmt.Sprintf("skills/direct/%s/%s.tar.gz", slug, sha),
		PackageSize: int64(len(md) + len(cfg)),
		Data:        append(append([]byte(nil), md...), cfg...),
	}, nil
}

func (p *fakePackager) PrepareCatalogFromDir(
	ctx context.Context,
	dir, repoIdentity string,
) (*extensionsvc.PreparedSkill, error) {
	prepared, err := p.PrepareFromDir(ctx, dir)
	if err != nil {
		return nil, err
	}
	p.catalogIdentities = append(p.catalogIdentities, repoIdentity)
	identityHash := sha256.Sum256([]byte(repoIdentity))
	prepared.StorageKey = fmt.Sprintf(
		"skills/catalog/%x/%s.tar.gz",
		identityHash,
		prepared.ContentSha,
	)
	return prepared, nil
}

func (p *fakePackager) StorePrepared(
	_ context.Context,
	prepared *extensionsvc.PreparedSkill,
) (*extensionsvc.PackagedSkill, error) {
	return &extensionsvc.PackagedSkill{
		Slug:        prepared.Slug,
		DisplayName: prepared.DisplayName,
		Description: prepared.Description,
		ContentSha:  prepared.ContentSha,
		StorageKey:  prepared.StorageKey,
		PackageSize: prepared.PackageSize,
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
