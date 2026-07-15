package goalloop

import (
	"context"
	"errors"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	workerspecdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	agentpodsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	workerspecsvc "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

var (
	ErrNotFound             = errors.New("goal loop not found")
	ErrInvalidInput         = errors.New("invalid goal loop input")
	ErrInvalidState         = errors.New("goal loop state does not allow this action")
	ErrExecutionUnavailable = errors.New("goal loop execution is not configured")
	ErrVerificationPending  = errors.New("goal loop verification is not ready")
)

type PodCreator interface {
	CreatePod(ctx context.Context, req *agentpodsvc.OrchestrateCreatePodRequest) (*agentpodsvc.OrchestrateCreatePodResult, error)
}

type PodLookup interface {
	GetPod(ctx context.Context, podKey string) (*agentpod.Pod, error)
}

type PodTerminator interface {
	TerminatePod(ctx context.Context, podKey string) error
}

type VerificationDispatcher interface {
	SendRunVerification(ctx context.Context, runnerID int64, cmd *runnerv1.RunVerificationCommand) error
}

type PromptDispatcher interface {
	Enabled() bool
	EnqueueSendPrompt(
		ctx context.Context,
		organizationID, runnerID int64,
		podKey, commandID, prompt string,
		ttl time.Duration,
	) error
}

type WorkerSpecSnapshotLoader interface {
	GetByID(ctx context.Context, organizationID, snapshotID int64) (workerspecdomain.Snapshot, error)
	ListByOrganization(ctx context.Context, organizationID int64) ([]workerspecdomain.Snapshot, error)
}

type WorkerTypeSnapshotValidator interface {
	ValidateWorkerTypeSnapshot(
		context.Context,
		workerspecsvc.Scope,
		workerspecdomain.WorkerType,
	) error
}

type Service struct {
	repo               domain.Repository
	podCreator         PodCreator
	podLookup          PodLookup
	podTerminator      PodTerminator
	verificationSender VerificationDispatcher
	promptSender       PromptDispatcher
	workerSpecs        WorkerSpecSnapshotLoader
	workerTypes        WorkerTypeSnapshotValidator
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetExecutionDependencies(
	podCreator PodCreator,
	podLookup PodLookup,
	podTerminator PodTerminator,
) {
	s.podCreator = podCreator
	s.podLookup = podLookup
	s.podTerminator = podTerminator
}

func (s *Service) SetVerificationDispatcher(sender VerificationDispatcher) {
	s.verificationSender = sender
}

func (s *Service) SetPromptDispatcher(sender PromptDispatcher) {
	s.promptSender = sender
}

func (s *Service) SetWorkerSpecSnapshotLoader(loader WorkerSpecSnapshotLoader) {
	s.workerSpecs = loader
}

func (s *Service) SetWorkerTypeSnapshotValidator(validator WorkerTypeSnapshotValidator) {
	s.workerTypes = validator
}

func (s *Service) ValidateExecutionReady() error {
	if !s.executionReady() {
		return ErrExecutionUnavailable
	}
	return nil
}

func (s *Service) GetBySlug(ctx context.Context, orgID int64, slug string) (*domain.GoalLoop, error) {
	loop, err := s.repo.GetBySlug(ctx, orgID, slug)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, ErrNotFound
	}
	return loop, err
}

func (s *Service) List(ctx context.Context, filter domain.ListFilter) ([]*domain.GoalLoop, int64, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) ListWorkerSnapshots(
	ctx context.Context,
	organizationID, userID int64,
) ([]workerspecdomain.Snapshot, error) {
	if s.workerSpecs == nil || s.workerTypes == nil {
		return nil, ErrExecutionUnavailable
	}
	snapshots, err := s.workerSpecs.ListByOrganization(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	available := make([]workerspecdomain.Snapshot, 0, len(snapshots))
	scope := workerspecsvc.Scope{OrgID: organizationID, UserID: userID}
	for _, snapshot := range snapshots {
		err := s.workerTypes.ValidateWorkerTypeSnapshot(
			ctx,
			scope,
			snapshot.Spec.Runtime.WorkerType,
		)
		if errors.Is(err, workercreation.ErrWorkerTypeDefinitionChanged) {
			continue
		}
		if err != nil {
			return nil, err
		}
		available = append(available, snapshot)
	}
	return available, nil
}

func (s *Service) ValidateWorkerSnapshotForExecution(
	ctx context.Context,
	organizationID, userID, snapshotID int64,
) error {
	if s.workerSpecs == nil || s.workerTypes == nil {
		return ErrExecutionUnavailable
	}
	snapshot, err := s.workerSpecs.GetByID(ctx, organizationID, snapshotID)
	if errors.Is(err, workerspecdomain.ErrNotFound) {
		return ErrInvalidInput
	}
	if err != nil {
		return err
	}
	if snapshot.OrganizationID != organizationID {
		return ErrInvalidInput
	}
	err = s.workerTypes.ValidateWorkerTypeSnapshot(
		ctx,
		workerspecsvc.Scope{OrgID: organizationID, UserID: userID},
		snapshot.Spec.Runtime.WorkerType,
	)
	if errors.Is(err, workercreation.ErrWorkerTypeDefinitionChanged) {
		return ErrInvalidInput
	}
	return err
}
