package skill

import (
	"context"
	"fmt"
	"strings"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

const operatorCatalogIdentityPrefix = "operator-catalog/"

type PlatformCatalogService struct {
	store    skilldom.Repository
	packager SkillPackagerBridge
}

func NewPlatformCatalogService(
	store skilldom.Repository,
	packager SkillPackagerBridge,
) *PlatformCatalogService {
	if store == nil || packager == nil {
		return nil
	}
	return &PlatformCatalogService{
		store:    store,
		packager: packager,
	}
}

func (s *PlatformCatalogService) EnsurePlatformSkill(
	ctx context.Context,
	req *EnsurePlatformSkillRequest,
) (*skilldom.Skill, bool, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, false, ErrNameRequired
	}
	if strings.TrimSpace(req.Instructions) == "" {
		return nil, false, ErrInstructionsRequired
	}
	tags, err := ValidateTags(req.Tags)
	if err != nil {
		return nil, false, err
	}
	slug := strings.TrimSpace(req.Slug)
	if err := slugkit.ValidateIdentifier("skills.slug", slug); err != nil {
		return nil, false, err
	}
	request := *req
	request.Slug = slug
	files, err := renderSkillFiles(
		slug,
		req.Name,
		req.Description,
		req.License,
		req.Instructions,
		tags,
	)
	if err != nil {
		return nil, false, err
	}
	prepared, err := s.prepare(ctx, slug, files)
	if err != nil {
		return nil, false, err
	}
	return s.publish(ctx, &request, tags, prepared)
}

func (s *PlatformCatalogService) prepare(
	ctx context.Context,
	slug string,
	files []gitops.FileChange,
) (*extensionsvc.PreparedSkill, error) {
	dir, cleanup, err := materializeFileChanges(files)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	prepared, err := s.packager.PrepareCatalogFromDir(
		ctx,
		dir,
		operatorCatalogIdentityPrefix+slug,
	)
	if err != nil {
		return nil, fmt.Errorf("skill: package operator catalog: %w", err)
	}
	return prepared, nil
}
