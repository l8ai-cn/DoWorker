package coordinator

import "context"

type ProjectListFilter struct {
	OrganizationID int64
	Enabled        *bool
}

type Repository interface {
	CreateProject(ctx context.Context, project *Project) error
	GetProject(ctx context.Context, orgID, id int64) (*Project, error)
	GetProjectBySlug(ctx context.Context, orgID int64, slug string) (*Project, error)
	ListProjects(ctx context.Context, filter *ProjectListFilter) ([]*Project, error)
	ListEnabledProjects(ctx context.Context) ([]*Project, error)
	UpdateProject(ctx context.Context, orgID, id int64, updates map[string]any) error
	DeleteProject(ctx context.Context, orgID, id int64) error

	GetLinkByExternalID(ctx context.Context, orgID int64, platformType, externalID string) (*TicketExternalLink, error)
	CreateLink(ctx context.Context, link *TicketExternalLink) error

	CreateExecution(ctx context.Context, execution *Execution) error
	GetExecution(ctx context.Context, id int64) (*Execution, error)
	GetExecutionByPodKey(ctx context.Context, podKey string) (*Execution, error)
	ListExecutions(ctx context.Context, projectID int64, limit int) ([]*Execution, error)
	UpdateExecution(ctx context.Context, id int64, updates map[string]any) error
	CountActiveExecutions(ctx context.Context, projectID int64) (int64, error)
}
