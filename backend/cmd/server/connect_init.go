package main

import (
	"net/http"

	"connectrpc.com/connect"

	adminconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/admin"
	adminauthconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/admin/auth"
	promocodeadminconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/admin/promocode"
	ssoadminconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/admin/sso"
	subscriptionadminconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/admin/subscription"
	supportticketadminconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/admin/support_ticket"
	agentconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/agent"
	apikeyconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/apikey"
	bindingconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/binding"
	blockstoreconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/blockstore"
	channelconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/channel"
	envbundleconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/env_bundle"
	eventsconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/events"
	executionclusterconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/executioncluster"
	extensionconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/extension"
	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	meshconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/mesh"
	orgconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/org"
	promocodeconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/promocode"
	repositoryconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/repository"
	supportticketconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/support_ticket"
	ticketconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/ticket"
	ticketrelationsconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/ticket_relations"
	userconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/user"
	usercredentialconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/user_credential"
	v1 "github.com/anthropics/agentsmesh/backend/internal/api/rest/v1"
	"github.com/anthropics/agentsmesh/backend/internal/config"
)

// defaultConnectHandlerOptions returns the HandlerOption set applied to
// every Connect handler. The auth interceptor mirrors REST's
// `middleware.AuthMiddleware`: it parses `Authorization: Bearer …`,
// validates the JWT against `cfg.JWT.Secret`, and injects the resulting
// `*middleware.TenantContext` (with UserID only — org scoping is the
// service handler's job) into the request context.
//
// Per-service Mount functions accept `...connect.HandlerOption` and
// must thread these through:
//
//	func Mount(mux *http.ServeMux, srv *Server, opts ...connect.HandlerOption) {
//	    path, h := fooconnect.NewFooServiceHandler(srv, opts...)
//	    mux.Handle(path, h)
//	}
//
// Callers in wrapWithConnect wire it as `Mount(mux, srv, defaults...)`.
func defaultConnectHandlerOptions(svc *serviceContainer) []connect.HandlerOption {
	return []connect.HandlerOption{
		connect.WithInterceptors(
			interceptors.NewAuthInterceptor(
				svc.auth.AccessTokenManager(),
				svc.auth.AccessTokenAudience(),
			),
		),
	}
}

// wrapWithConnect returns a top-level handler that prefers Connect for
// `/proto.*` paths and falls through to the Gin REST router for
// everything else. Per-service Mount calls registered onto connectMux
// here pick up the default HandlerOptions (auth interceptor, …); the
// REST router is untouched.
//
// `rest` already has every optional dependency (PodCoordinator,
// VersionChecker, etc.) threaded through `v1.Services` — we pass it in
// alongside `svc` so Connect handlers can share the same wiring.
func wrapWithConnect(cfg *config.Config, svc *serviceContainer, rest *v1.Services, restHandler http.Handler) http.Handler {
	connectMux := http.NewServeMux()
	opts := defaultConnectHandlerOptions(svc)

	mountConnectServices(connectMux, svc, rest, cfg, opts)

	combined := routeConnectOrREST(withConnectTracing(connectMux), restHandler)
	return withBrowserCORS(cfg.Server.CORSAllowedOrigins, combined)
}

// mountConnectServices is the seam each per-service migration PR adds
// to. Specialist PRs insert one line per service.
func mountConnectServices(mux *http.ServeMux, svc *serviceContainer, rest *v1.Services, cfg *config.Config, opts []connect.HandlerOption) {
	mountAgentWorkbenchService(mux, svc, rest)
	extensionSrv := extensionconnect.NewServer(svc.extension, svc.org)
	extensionconnect.MountMarket(mux, extensionconnect.NewMarketServer(extensionSrv), opts...)
	extensionconnect.MountRepoSkill(mux, extensionconnect.NewRepoSkillServer(extensionSrv), opts...)
	extensionconnect.MountRepoMcp(mux, extensionconnect.NewRepoMcpServer(extensionSrv), opts...)
	repositoryconnect.Mount(mux, repositoryconnect.NewServer(
		svc.repository, svc.org,
		repositoryconnect.WithBillingService(svc.billing),
	), opts...)
	apikeyconnect.Mount(mux, apikeyconnect.NewServer(svc.apikey, svc.org), opts...)
	bindingconnect.Mount(mux, bindingconnect.NewServer(svc.binding, svc.org), opts...)
	mountOrchestrationResourceService(mux, svc, opts)
	if svc.blockstore != nil {
		blockstoreconnect.Mount(mux, blockstoreconnect.NewServer(svc.blockstore, svc.org), opts...)
	}
	orgconnect.Mount(mux, orgconnect.NewServer(svc.org, svc.user), opts...)
	ticketrelationsconnect.Mount(mux, ticketrelationsconnect.NewServer(svc.ticket, svc.org), opts...)
	channelconnect.Mount(mux, channelconnect.NewServer(svc.channel, svc.ticket, svc.org), opts...)
	ticketconnect.Mount(mux, ticketconnect.NewServer(svc.ticket, svc.org), opts...)
	meshconnect.Mount(mux, meshconnect.NewServer(svc.mesh, svc.ticket, svc.org), opts...)
	if svc.message != nil {
		meshconnect.MountMessages(mux, meshconnect.NewMessageServer(svc.message, svc.org), opts...)
	}
	if rest != nil && rest.Hub != nil {
		eventsconnect.Mount(mux, eventsconnect.NewServer(rest.Hub, svc.org), opts...)
	}
	if svc.executionCluster != nil {
		executionclusterconnect.Mount(mux, executionclusterconnect.NewServer(svc.executionCluster, svc.org), opts...)
	}
	mountRunnerService(mux, svc, rest, cfg, opts)
	mountPodService(mux, svc, rest, cfg, opts)
	mountAgentPodSettingsService(mux, svc, opts)
	mountAIResourceService(mux, svc, opts)
	usercredentialconnect.Mount(mux, usercredentialconnect.NewServer(svc.user), opts...)
	userconnect.Mount(mux, userconnect.NewServer(svc.user, svc.org), opts...)
	agentconnect.Mount(mux, agentconnect.NewServer(
		svc.agentSvc, svc.envBundle, svc.userConfig, svc.org, svc.workerDefinitions,
	), opts...)
	if svc.envBundle != nil {
		envbundleconnect.Mount(mux, envbundleconnect.NewServer(svc.envBundle), opts...)
	}
	mountBillingService(mux, svc, opts)
	mountInvitationService(mux, svc, opts)
	promocodeconnect.Mount(mux, promocodeconnect.NewServer(svc.promoCode, svc.org), opts...)
	supportticketconnect.Mount(mux, supportticketconnect.NewServer(svc.supportTicket), opts...)
	mountSSOService(mux, svc)
	mountAuthService(mux, svc, rest, cfg, opts)
	mountGrantService(mux, svc, opts)
	mountFileService(mux, svc, opts)
	mountTokenUsageService(mux, svc, opts)
	mountAutopilotService(mux, svc, rest, opts)
	mountNotificationService(mux, svc, opts)
	mountWorkflowService(mux, svc, rest, opts)
	mountGoalLoopService(mux, svc, opts)
	mountKnowledgeBaseService(mux, svc, opts)
	mountLicenseService(mux, svc, opts)
	mountAdminServices(mux, svc, rest, cfg, opts)
}

// mountAdminServices wires the platform-admin Connect surface. Every
// handler in this group internally calls interceptors.ResolveSystemAdmin
// against svc.adminDB to mirror REST's AdminMiddleware. Skips silently
// when svc.admin is nil (admin disabled by config) — same gate the REST
// router applies at router.go:202.
//
// rest.RelayManager (optional) threads through WithRelayManager so the
// 4 relay RPCs work in deployments that wire the relay subsystem. Same
// nil-guard pattern as REST's admin/routes.go:70.
func mountAdminServices(mux *http.ServeMux, svc *serviceContainer, rest *v1.Services, cfg *config.Config, opts []connect.HandlerOption) {
	if svc.admin == nil {
		return
	}
	adminOpts := []adminconnect.Option{}
	if rest != nil && rest.RelayManager != nil {
		adminOpts = append(adminOpts, adminconnect.WithRelayManager(rest.RelayManager))
	}
	if rest != nil && rest.Message != nil {
		adminOpts = append(adminOpts, adminconnect.WithMessageService(rest.Message))
	}
	if rest != nil && rest.Expert != nil {
		adminOpts = append(adminOpts, adminconnect.WithExpertService(rest.Expert))
	}
	adminconnect.Mount(mux, adminconnect.NewServer(svc.admin, svc.adminDB, adminOpts...), opts...)
	promocodeadminconnect.Mount(mux, promocodeadminconnect.NewServer(svc.admin, svc.adminDB), opts...)
	if svc.billing != nil {
		subscriptionadminconnect.Mount(mux, subscriptionadminconnect.NewServer(svc.admin, svc.billing, svc.adminDB), opts...)
	}
	if svc.sso != nil {
		ssoadminconnect.Mount(
			mux,
			ssoadminconnect.NewServer(svc.sso, svc.admin, svc.adminDB),
			opts...,
		)
	}
	if svc.supportTicket != nil {
		supportticketadminconnect.Mount(
			mux,
			supportticketadminconnect.NewServer(svc.supportTicket, svc.admin, svc.adminDB),
			opts...,
		)
	}

	// AdminAuthService.Login is PUBLIC (no auth interceptor — caller
	// doesn't hold a bearer yet); session lookup goes behind opts.
	if svc.auth != nil {
		adminauthconnect.MountLogin(mux, adminauthconnect.NewLoginServer(svc.auth, cfg))
		adminauthconnect.MountSession(mux, adminauthconnect.NewSessionServer(svc.adminDB), opts...)
	}
}
