package coordinator

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	coordinatordom "github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

var ErrInvalidName = errors.New("coordinator: project name is required")

type CreateProjectRequest struct {
	OrganizationID      int64
	RepositoryID        int64
	Name                string
	PlatformType        string
	SourceType          string
	LabelFilter         []string
	ClaimPolicy         coordinatordom.ClaimPolicy
	AgentSlug           string
	ScanIntervalSeconds int
	MaxConcurrent       int
	CreatedByID         int64
}

func (s *Service) CreateProject(ctx context.Context, req *CreateProjectRequest) (*coordinatordom.Project, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, ErrInvalidName
	}
	slug, err := slugkit.GenerateUnique(ctx, req.Name, slugkit.FromExistsCheck(func(ctx context.Context, candidate string) (bool, error) {
		_, err := s.store.GetProjectBySlug(ctx, req.OrganizationID, candidate)
		if errors.Is(err, coordinatordom.ErrNotFound) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return true, nil
	}))
	if err != nil {
		return nil, err
	}

	policyJSON, err := json.Marshal(req.ClaimPolicy)
	if err != nil {
		return nil, err
	}

	project := &coordinatordom.Project{
		OrganizationID:      req.OrganizationID,
		RepositoryID:        req.RepositoryID,
		Slug:                slug,
		Name:                req.Name,
		PlatformType:        firstNonEmpty(req.PlatformType, coordinatordom.PlatformTypeCNB),
		SourceType:          firstNonEmpty(req.SourceType, coordinatordom.SourceTypeIssues),
		LabelFilter:         req.LabelFilter,
		ClaimPolicy:         policyJSON,
		AgentSlug:           firstNonEmpty(req.AgentSlug, "do-agent"),
		ScanIntervalSeconds: defaultInt(req.ScanIntervalSeconds, 300),
		MaxConcurrent:       defaultInt(req.MaxConcurrent, 1),
		Enabled:             true,
		CreatedByID:         req.CreatedByID,
	}
	if err := s.store.CreateProject(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *Service) ListProjects(ctx context.Context, orgID int64) ([]*coordinatordom.Project, error) {
	return s.store.ListProjects(ctx, &coordinatordom.ProjectListFilter{OrganizationID: orgID})
}

func (s *Service) GetProject(ctx context.Context, orgID, id int64) (*coordinatordom.Project, error) {
	return s.store.GetProject(ctx, orgID, id)
}

func (s *Service) UpdateProject(ctx context.Context, orgID, id int64, updates map[string]any) error {
	if _, err := s.store.GetProject(ctx, orgID, id); err != nil {
		return err
	}
	return s.store.UpdateProject(ctx, orgID, id, updates)
}

func (s *Service) DeleteProject(ctx context.Context, orgID, id int64) error {
	if _, err := s.store.GetProject(ctx, orgID, id); err != nil {
		return err
	}
	return s.store.DeleteProject(ctx, orgID, id)
}

func (s *Service) ListExecutions(ctx context.Context, orgID, projectID int64, limit int) ([]*coordinatordom.Execution, error) {
	if _, err := s.store.GetProject(ctx, orgID, projectID); err != nil {
		return nil, err
	}
	return s.store.ListExecutions(ctx, projectID, limit)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func defaultInt(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
