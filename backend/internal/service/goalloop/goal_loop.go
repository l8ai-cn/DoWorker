package goalloop

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	workerspecdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	agentpodsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
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

type WorkerSpecSnapshotLoader interface {
	GetByID(ctx context.Context, organizationID, snapshotID int64) (workerspecdomain.Snapshot, error)
}

type Service struct {
	repo               domain.Repository
	podCreator         PodCreator
	podLookup          PodLookup
	podTerminator      PodTerminator
	autopilot          *agentpodsvc.AutopilotControllerService
	verificationSender VerificationDispatcher
	workerSpecs        WorkerSpecSnapshotLoader
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetExecutionDependencies(
	podCreator PodCreator,
	podLookup PodLookup,
	podTerminator PodTerminator,
	autopilot *agentpodsvc.AutopilotControllerService,
) {
	s.podCreator = podCreator
	s.podLookup = podLookup
	s.podTerminator = podTerminator
	s.autopilot = autopilot
}

func (s *Service) SetVerificationDispatcher(sender VerificationDispatcher) {
	s.verificationSender = sender
}

func (s *Service) SetWorkerSpecSnapshotLoader(loader WorkerSpecSnapshotLoader) {
	s.workerSpecs = loader
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
