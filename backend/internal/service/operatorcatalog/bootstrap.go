package operatorcatalog

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	expertdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/expert"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/expertmarket"
	skilldom "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	expertsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/expert"
	skillsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/skill"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

var ErrCatalogConflict = errors.New("operator catalog conflicts with existing data")

type SkillEnsurer interface {
	EnsurePlatformSkill(
		context.Context,
		*skillsvc.EnsurePlatformSkillRequest,
	) (*skilldom.Skill, bool, error)
}

type ExpertPublisher interface {
	GetBySlug(context.Context, int64, string) (*expertdom.Expert, error)
	Create(context.Context, *expertsvc.CreateExpertRequest) (*expertdom.Expert, error)
	RebindWorkerSpecSnapshot(
		context.Context,
		int64,
		int64,
		int64,
	) (*expertdom.Expert, error)
	SubmitMarketApplication(
		context.Context,
		expertsvc.SubmitMarketApplicationRequest,
	) (*expertsvc.MarketSubmission, error)
	GetPublishedMarketApplication(
		context.Context,
		string,
	) (*expertsvc.PublishedMarketApplication, error)
	ListPublisherMarketReleases(
		context.Context,
		int64,
		int,
		int,
	) ([]expertmarket.Release, int64, error)
	ApproveMarketRelease(
		context.Context,
		expertsvc.ReviewMarketReleaseRequest,
	) (*expertmarket.Release, error)
}

type WorkerPreparer interface {
	Revision() string
	Prepare(
		context.Context,
		specservice.Scope,
		workercreation.Draft,
	) (workercreation.Prepared, error)
}

type SnapshotStore interface {
	Create(
		context.Context,
		specservice.ResolvedSnapshot,
	) (specdomain.Snapshot, error)
	GetByID(context.Context, int64, int64) (specdomain.Snapshot, error)
	Delete(context.Context, int64, int64) error
}

type Bootstrapper struct {
	skills    SkillEnsurer
	experts   ExpertPublisher
	workers   WorkerPreparer
	snapshots SnapshotStore
	artifacts DependencyArtifactStore
}

type BootstrapRequest struct {
	OrganizationID   int64
	OrganizationSlug slugkit.Slug
	PublisherUserID  int64
	ReviewerUserID   int64
	ModelResourceID  int64
	RuntimeImageID   int64
}

type BootstrapResult struct {
	CreatedSkills  int
	CreatedExperts int
	Published      int
}

func NewBootstrapper(
	skills SkillEnsurer,
	experts ExpertPublisher,
	workers WorkerPreparer,
	snapshots SnapshotStore,
	artifacts DependencyArtifactStore,
) *Bootstrapper {
	return &Bootstrapper{
		skills: skills, experts: experts, workers: workers, snapshots: snapshots,
		artifacts: artifacts,
	}
}

func (bootstrapper *Bootstrapper) Run(
	ctx context.Context,
	request BootstrapRequest,
) (BootstrapResult, error) {
	if err := bootstrapper.validate(request); err != nil {
		return BootstrapResult{}, err
	}
	skillDefinitions, err := Skills()
	if err != nil {
		return BootstrapResult{}, err
	}
	rows := make(map[string]*skilldom.Skill, len(skillDefinitions))
	result := BootstrapResult{}
	for _, definition := range skillDefinitions {
		row, created, ensureErr := bootstrapper.skills.EnsurePlatformSkill(
			ctx,
			platformSkillRequest(request, definition),
		)
		if ensureErr != nil {
			return result, fmt.Errorf("ensure skill %s: %w", definition.Slug, ensureErr)
		}
		rows[definition.Slug] = row
		if created {
			result.CreatedSkills++
		}
	}
	for _, definition := range Experts() {
		created, published, ensureErr := bootstrapper.ensureExpert(
			ctx,
			request,
			definition,
			rows,
		)
		if ensureErr != nil {
			return result, fmt.Errorf("ensure expert %s: %w", definition.Slug, ensureErr)
		}
		if created {
			result.CreatedExperts++
		}
		if published {
			result.Published++
		}
	}
	return result, nil
}

func (bootstrapper *Bootstrapper) validate(request BootstrapRequest) error {
	if bootstrapper == nil ||
		dependencyMissing(bootstrapper.skills) ||
		dependencyMissing(bootstrapper.experts) ||
		dependencyMissing(bootstrapper.workers) ||
		dependencyMissing(bootstrapper.snapshots) ||
		dependencyMissing(bootstrapper.artifacts) {
		return errors.New("operator catalog dependencies are incomplete")
	}
	if request.OrganizationID <= 0 ||
		slugkit.Validate(request.OrganizationSlug.String()) != nil ||
		request.PublisherUserID <= 0 ||
		request.ReviewerUserID <= 0 || request.ModelResourceID <= 0 ||
		request.RuntimeImageID <= 0 {
		return errors.New("operator catalog bootstrap identifiers must be positive")
	}
	return nil
}

func dependencyMissing(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	return reflected.Kind() == reflect.Pointer && reflected.IsNil()
}

func platformSkillRequest(
	request BootstrapRequest,
	definition SkillDefinition,
) *skillsvc.EnsurePlatformSkillRequest {
	return &skillsvc.EnsurePlatformSkillRequest{
		UserID:       request.PublisherUserID,
		Slug:         definition.Slug,
		Name:         definition.Name,
		Description:  definition.Description,
		License:      definition.License,
		Instructions: definition.Instructions,
		Tags:         definition.Tags,
	}
}
