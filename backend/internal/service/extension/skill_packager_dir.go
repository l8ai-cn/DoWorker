package extension

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// PackageFromDir is the combined compatibility path for callers that do not
// coordinate package publication with a catalog transaction.
func (p *SkillPackager) PackageFromDir(ctx context.Context, dir string) (*PackagedSkill, error) {
	prepared, err := p.PrepareFromDir(ctx, dir)
	if err != nil {
		return nil, err
	}
	return p.StorePrepared(ctx, prepared)
}

func (p *SkillPackager) PrepareFromDir(_ context.Context, dir string) (*PreparedSkill, error) {
	if !fileExists(filepath.Join(dir, "SKILL.md")) {
		return nil, fmt.Errorf("SKILL.md not found in %s", dir)
	}
	if !dirExists(dir) {
		return nil, fmt.Errorf("skill dir %q not found", dir)
	}
	if _, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("skill dir: %w", err)
	}
	return p.prepareDir(dir)
}

func (p *SkillPackager) StorePrepared(
	ctx context.Context,
	prepared *PreparedSkill,
) (*PackagedSkill, error) {
	exists, err := p.storage.Exists(ctx, prepared.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing package: %w", err)
	}
	created := !exists
	if created {
		_, err = p.storage.Upload(
			ctx,
			prepared.StorageKey,
			bytes.NewReader(prepared.Data),
			prepared.PackageSize,
			"application/gzip",
		)
		if err != nil {
			slog.ErrorContext(
				ctx,
				"failed to upload skill package",
				"slug", prepared.Slug,
				"storage_key", prepared.StorageKey,
				"error", err,
			)
			return nil, fmt.Errorf("failed to upload: %w", err)
		}
	}
	slog.InfoContext(
		ctx,
		"skill package ready",
		"slug", prepared.Slug,
		"content_sha", prepared.ContentSha,
		"package_size", prepared.PackageSize,
		"created", created,
	)
	return &PackagedSkill{
		Slug:        prepared.Slug,
		DisplayName: prepared.DisplayName,
		Description: prepared.Description,
		ContentSha:  prepared.ContentSha,
		StorageKey:  prepared.StorageKey,
		PackageSize: prepared.PackageSize,
		Created:     created,
	}, nil
}

func (p *SkillPackager) DeletePackage(ctx context.Context, storageKey string) error {
	if err := p.storage.Delete(ctx, storageKey); err != nil {
		return fmt.Errorf("delete skill package %q: %w", storageKey, err)
	}
	return nil
}
