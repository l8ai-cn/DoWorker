// Package podconnect hosts Connect-RPC handlers for the pod domain.
// Mirrors backend/internal/api/rest/v1/pod*.go but exposes the data plane
// via Connect (binary protobuf wire, see conventions.md §2.5). REST stays
// mounted in parallel; the migration runs dual-track until all 26 services
// have flipped.
//
// Streaming endpoints (terminal data plane, pod events) intentionally stay
// on Relay/WebSocket — this migration is unary RPC only.
//
// Handler shape follows runbook §3:
//   - ResolveOrgScope reads org_slug + injects TenantContext.
//   - Single-entity get/create/update return the entity directly.
//   - List responses follow {items, total, limit, offset}.
//   - CreatePod keeps {pod, warning?} envelope locked by 986a38ca6 (PR #340).
//   - Errors map to Connect codes (conventions §10).
package podconnect

import (
	"context"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/grant"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	itemservice "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	"github.com/anthropics/agentsmesh/backend/internal/service/geo"
	grantservice "github.com/anthropics/agentsmesh/backend/internal/service/grant"
	"github.com/anthropics/agentsmesh/backend/internal/service/relay"
	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/policy"
)

// ServiceName mirrors proto.pod.v1.PodService exactly — Connect derives the
// URL from `<package>.<Service>` (conventions §1, §12).
const ServiceName = "proto.pod.v1.PodService"

const (
	ListPodsProcedure                  = "/" + ServiceName + "/ListPods"
	GetPodProcedure                    = "/" + ServiceName + "/GetPod"
	CreatePodProcedure                 = "/" + ServiceName + "/CreatePod"
	TerminatePodProcedure              = "/" + ServiceName + "/TerminatePod"
	UpdatePodAliasProcedure            = "/" + ServiceName + "/UpdatePodAlias"
	UpdatePodPerpetualProcedure        = "/" + ServiceName + "/UpdatePodPerpetual"
	UpdatePodPreviewConfigProcedure    = "/" + ServiceName + "/UpdatePodPreviewConfig"
	GetMobileAccessDescriptorProcedure = "/" + ServiceName + "/GetMobileAccessDescriptor"
	GetPodConnectionProcedure          = "/" + ServiceName + "/GetPodConnection"
	SendPodPromptProcedure             = "/" + ServiceName + "/SendPodPrompt"
	ListPodsByTicketProcedure          = "/" + ServiceName + "/ListPodsByTicket"
	ListWorkerCreateOptionsProcedure   = "/" + ServiceName + "/ListWorkerCreateOptions"
	PreflightWorkerProcedure           = "/" + ServiceName + "/PreflightWorker"
	FillWorkerDraftProcedure           = "/" + ServiceName + "/FillWorkerDraft"
)

type WorkerCreationAPI interface {
	ListOptions(
		context.Context,
		specservice.Scope,
		workercreation.OptionsFilter,
	) (workercreation.CreateOptions, error)
	Preflight(
		context.Context,
		specservice.Scope,
		workercreation.Draft,
	) (workercreation.PreflightResult, error)
}

type WorkerDraftFiller interface {
	Fill(
		context.Context,
		specservice.Scope,
		string,
		int64,
		*workercreation.Draft,
	) (workercreation.FillResult, error)
}

type WorkerSpecSnapshotLoader interface {
	GetByID(
		context.Context,
		int64,
		int64,
	) (workerspec.Snapshot, error)
	GetByIDs(
		context.Context,
		int64,
		[]int64,
	) ([]workerspec.Snapshot, error)
}

type relayTokenGenerator interface {
	GenerateToken(string, int64, int64, int64, time.Duration) (string, error)
}

// Server implements PodService. Fields mirror PodHandler / PodConnectHandler
// in v1/pods.go and v1/pod_relay_connect.go, threaded through cmd/server
// wiring at mount time. Streaming endpoints (terminal data plane) intentionally
// stay on Relay/WebSocket — Connect handles unary control plane only.
type Server struct {
	podSvc            *agentpod.PodService
	orgSvc            middleware.OrganizationService
	orchestrator      *agentpod.PodOrchestrator
	podCoordinator    *runner.PodCoordinator
	commandSender     runner.RunnerCommandSender
	relayManager      *relay.Manager
	tokenGenerator    relayTokenGenerator
	geoResolver       geo.Resolver
	grantSvc          *grantservice.Service
	eventBus          *eventbus.EventBus
	workerCreation    WorkerCreationAPI
	workerDraftFiller WorkerDraftFiller
	workerSpecs       WorkerSpecSnapshotLoader
	conversationItems itemservice.PositionedAppender
	mobileBaseURL     string
}

// NewServer constructs a Server. Optional dependencies can be left nil; the
// corresponding handlers degrade gracefully (CodeUnavailable for missing
// command-sender, etc.).
func NewServer(
	podSvc *agentpod.PodService,
	orgSvc middleware.OrganizationService,
	opts ...Option,
) *Server {
	s := &Server{podSvc: podSvc, orgSvc: orgSvc}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option configures a Server. Mirrors the PodHandlerOption pattern in
// v1/pods.go so wiring stays parallel between Gin and Connect.
type Option func(*Server)

func WithOrchestrator(o *agentpod.PodOrchestrator) Option {
	return func(s *Server) { s.orchestrator = o }
}

func WithPodCoordinator(pc *runner.PodCoordinator) Option {
	return func(s *Server) { s.podCoordinator = pc }
}

func WithCommandSender(cs runner.RunnerCommandSender) Option {
	return func(s *Server) { s.commandSender = cs }
}

func WithRelayManager(rm *relay.Manager) Option {
	return func(s *Server) { s.relayManager = rm }
}

func WithTokenGenerator(tg *relay.TokenGenerator) Option {
	return func(s *Server) { s.tokenGenerator = tg }
}

func WithGeoResolver(gr geo.Resolver) Option {
	return func(s *Server) { s.geoResolver = gr }
}

func WithGrantService(gs *grantservice.Service) Option {
	return func(s *Server) { s.grantSvc = gs }
}

func WithEventBus(eb *eventbus.EventBus) Option {
	return func(s *Server) { s.eventBus = eb }
}

func WithWorkerCreation(service WorkerCreationAPI) Option {
	return func(server *Server) { server.workerCreation = service }
}

func WithWorkerDraftFiller(filler WorkerDraftFiller) Option {
	return func(server *Server) { server.workerDraftFiller = filler }
}

func WithWorkerSpecSnapshotLoader(loader WorkerSpecSnapshotLoader) Option {
	return func(server *Server) { server.workerSpecs = loader }
}

func WithConversationItems(items itemservice.PositionedAppender) Option {
	return func(server *Server) { server.conversationItems = items }
}

func WithMobileBaseURL(baseURL string) Option {
	return func(server *Server) { server.mobileBaseURL = baseURL }
}

// podResourceWithGrants mirrors PodHandler.podResourceWithGrants (v1/pod_relay_connect.go:56).
// Builds a policy.ResourceContext that respects per-user grants on the pod.
func (s *Server) podResourceWithGrants(ctx context.Context, podKey string, orgID, createdByID int64) policy.ResourceContext {
	rc := policy.PodResource(orgID, createdByID)
	if s.grantSvc == nil {
		return rc
	}
	if ids, err := s.grantSvc.GetGrantedUserIDs(ctx, grant.TypePod, podKey); err == nil && len(ids) > 0 {
		return rc.WithGrants(ids)
	}
	return rc
}
