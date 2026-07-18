package coordinator

import (
	"context"
	"log/slog"

	coordinatordom "github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"
	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	ticketDomain "github.com/anthropics/agentsmesh/backend/internal/domain/ticket"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	ticketSvc "github.com/anthropics/agentsmesh/backend/internal/service/ticket"
)

// TicketService is the slice of the ticket service the coordinator drives: it
// materializes external issues as tickets and advances their status on feedback.
type TicketService interface {
	CreateTicket(ctx context.Context, req *ticketSvc.CreateTicketRequest) (*ticketDomain.Ticket, error)
	GetTicket(ctx context.Context, ticketID int64) (*ticketDomain.Ticket, error)
	UpdateStatus(ctx context.Context, ticketID int64, status string) error
	DeleteTicket(ctx context.Context, ticketID int64) error
}

// PodDispatcher is satisfied by *agentpod.PodOrchestrator.
type PodDispatcher interface {
	CreatePod(ctx context.Context, req *agentpodSvc.OrchestrateCreatePodRequest) (*agentpodSvc.OrchestrateCreatePodResult, error)
}

type PodTerminator interface {
	TerminatePod(ctx context.Context, podKey string) error
}

type RepoResolver interface {
	GetByID(ctx context.Context, id int64) (*gitprovider.Repository, error)
}

type TokenProvider interface {
	GetDecryptedProviderTokenByTypeAndURL(ctx context.Context, userID int64, providerType, baseURL string) (string, error)
}

// PlatformFactory builds a TaskPlatform (and resolves the repo slug) for a
// project, injecting the org-scoped provider credential.
type PlatformFactory interface {
	For(ctx context.Context, project *coordinatordom.Project) (TaskPlatform, string, error)
}

type Service struct {
	store         coordinatordom.Repository
	tickets       TicketService
	dispatch      PodDispatcher
	podTerminator PodTerminator
	platform      PlatformFactory
	runnerEnsurer *RunnerEnsurer
	logger        *slog.Logger
}

type Deps struct {
	Store         coordinatordom.Repository
	Tickets       TicketService
	Dispatch      PodDispatcher
	PodTerminator PodTerminator
	Platform      PlatformFactory
	RunnerEnsurer *RunnerEnsurer
	Logger        *slog.Logger
}

func NewService(deps Deps) *Service {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		store:         deps.Store,
		tickets:       deps.Tickets,
		dispatch:      deps.Dispatch,
		podTerminator: deps.PodTerminator,
		platform:      deps.Platform,
		runnerEnsurer: deps.RunnerEnsurer,
		logger:        logger.With("component", "coordinator"),
	}
}

func (s *Service) Store() coordinatordom.Repository { return s.store }
