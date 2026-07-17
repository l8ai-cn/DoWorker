package main

import (
	"net/http"

	"connectrpc.com/connect"

	agentpodsettingsconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/agentpod_settings"
	authconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/auth"
	autopilotconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/autopilot"
	billingconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/billing"
	fileconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/file"
	grantconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/grant"
	invitationconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/invitation"
	knowledgebaseconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/knowledgebase"
	licenseconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/license"
	notificationconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/notification"
	ssoconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/sso"
	tokenusageconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/token_usage"
	workflowconnect "github.com/anthropics/agentsmesh/backend/internal/api/connect/workflow"
	v1 "github.com/anthropics/agentsmesh/backend/internal/api/rest/v1"
	"github.com/anthropics/agentsmesh/backend/internal/config"
)

// mountAuthService wires both AuthService (PUBLIC — no auth interceptor)
// and AuthSessionService (auth-required — Logout). Per conventions §3.5
// exception #1, the user does not have a bearer token when hitting
// Login/Register/etc.; the act of authenticating IS to obtain one.
// AuthSessionService.Logout is the only RPC that requires the token
// (it revokes the caller's own bearer).
//
// REST handlers in /api/v1/auth/* stay mounted permanently in parallel —
// AuthManager.login/refresh/logout in the Rust auth crate still drives
// the stateful auth flow via REST, this Connect surface is the data-plane
// migration target for register/verify/forgot/reset call sites and a
// forward-compatible path for future flow migrations.
func mountAuthService(
	mux *http.ServeMux,
	svc *serviceContainer,
	rest *v1.Services,
	cfg *config.Config,
	opts []connect.HandlerOption,
) {
	srv := authconnect.NewServer(svc.auth, svc.user, svc.email, cfg)
	authconnect.MountPublic(mux, srv)
	authconnect.MountSession(
		mux,
		authconnect.NewSessionServer(svc.auth, rest.PreviewSessions),
		opts...,
	)
}

// mountSSOService wires the public SSOService (Discover + LdapAuth)
// onto the mux WITHOUT the auth interceptor — conventions §3.5
// exception #1. The user does not have a bearer token when they hit
// these RPCs; that is the goal of the SSO login flow.
//
// The OIDC/SAML browser-redirect endpoints (auth_sso_oidc.go,
// auth_sso_saml.go) stay on REST permanently — Connect's unary
// contract cannot return `Location:` redirects.
func mountSSOService(mux *http.ServeMux, svc *serviceContainer) {
	srv := ssoconnect.NewServer(svc.sso, svc.auth, svc.user)
	ssoconnect.MountPublic(mux, srv)
}

// mountInvitationService wires the auth-required InvitationService +
// UserInvitationService and the unauthenticated PublicInvitationService onto
// the same mux. The public service skips `opts` — the auth interceptor would
// reject every token-only lookup from /invite/[token] before the user signs
// in. The token IS the credential (single-use, opaque hex).
func mountInvitationService(mux *http.ServeMux, svc *serviceContainer, opts []connect.HandlerOption) {
	srv := invitationconnect.NewServer(
		svc.invitation, svc.org, svc.org, svc.user,
		invitationconnect.WithBillingService(svc.billing),
	)
	invitationconnect.Mount(mux, srv, opts...)
	invitationconnect.MountPublic(mux, invitationconnect.NewPublicServer(svc.invitation))
}

// mountBillingService wires both BillingService (auth-required, org-scoped)
// and BillingPublicService (no auth, no org_slug) onto the same mux.
//
// The public service intentionally mounts WITHOUT `opts` — the auth
// interceptor would reject every unauthenticated request from the landing
// page. The handler relies on conventions §3.5's "User-scoped /
// Platform-admin scoped" exception (no `ResolveOrgScope` call).
func mountBillingService(mux *http.ServeMux, svc *serviceContainer, opts []connect.HandlerOption) {
	billingconnect.Mount(mux, billingconnect.NewServer(svc.billing, svc.org), opts...)
	billingconnect.MountPublic(mux, billingconnect.NewPublicServer(svc.billing))
}

// mountAgentPodSettingsService wires the user-scoped AgentPod settings +
// AI provider Connect server. No optional deps — both services are wired
// unconditionally during service init.
func mountAgentPodSettingsService(mux *http.ServeMux, svc *serviceContainer, opts []connect.HandlerOption) {
	srv := agentpodsettingsconnect.NewServer(svc.agentpodSettings, svc.agentpodAIProvider)
	agentpodsettingsconnect.Mount(mux, srv, opts...)
}

// mountGrantService wires GrantService — covers pod / runner / repository
// grants under one Connect endpoint. Skips when the grant service is nil
// (test wiring); the REST router does the same in routes_pods.go.
func mountGrantService(mux *http.ServeMux, svc *serviceContainer, opts []connect.HandlerOption) {
	if svc.grant == nil {
		return
	}
	srv := grantconnect.NewServer(svc.grant, svc.org, svc.pod, svc.runner, svc.repository)
	grantconnect.Mount(mux, srv, opts...)
}

// mountFileService wires FileService — presigned upload URL generation.
// Skips when the file service is nil (storage backend not configured).
func mountFileService(mux *http.ServeMux, svc *serviceContainer, opts []connect.HandlerOption) {
	if svc.file == nil {
		return
	}
	srv := fileconnect.NewServer(svc.file, svc.org)
	fileconnect.Mount(mux, srv, opts...)
}

// mountTokenUsageService wires TokenUsageService — admin-only dashboard
// for token consumption analytics. Skips when nil.
func mountTokenUsageService(mux *http.ServeMux, svc *serviceContainer, opts []connect.HandlerOption) {
	if svc.tokenUsage == nil {
		return
	}
	srv := tokenusageconnect.NewServer(svc.tokenUsage, svc.org, svc.podSessionUsage)
	tokenusageconnect.Mount(mux, srv, opts...)
}

// mountAutopilotService wires AutopilotControllerService — CRUD + 6
// control actions + iterations. Reuses the GRPCCommandSender from
// rest.RunnerCommandSender (same instance the REST handler uses).
func mountAutopilotService(mux *http.ServeMux, svc *serviceContainer, rest *v1.Services, opts []connect.HandlerOption) {
	if svc.autopilot == nil {
		return
	}
	var cmdSender autopilotconnect.CommandSender
	if rest != nil && rest.PodCoordinator != nil {
		if s, ok := rest.PodCoordinator.GetCommandSender().(autopilotconnect.CommandSender); ok {
			cmdSender = s
		}
	}
	var apOpts []autopilotconnect.Option
	if rest != nil && rest.EventBus != nil {
		apOpts = append(apOpts, autopilotconnect.WithEventBus(rest.EventBus))
	}
	srv := autopilotconnect.NewServer(svc.autopilot, svc.org, svc.pod, cmdSender, apOpts...)
	autopilotconnect.Mount(mux, srv, opts...)
}

// mountNotificationService wires NotificationService — per-user notification
// preference CRUD inside an org. REST stays mounted at
// /api/v1/orgs/:slug/notifications/preferences for the dual-track window.
// Phase 2 (unread-count subscribe stream) stays on the websocket relay path
// because Connect's unary contract cannot model server-push.
func mountNotificationService(mux *http.ServeMux, svc *serviceContainer, opts []connect.HandlerOption) {
	if svc.notifPrefStore == nil {
		return
	}
	srv := notificationconnect.NewServer(svc.notifPrefStore, svc.org)
	notificationconnect.Mount(mux, srv, opts...)
}

// mountKnowledgeBaseService wires KnowledgeBaseService — org-scoped llm-wiki
// CRUD + agent mounts + repo browsing. Skips when the internal Gitea is not
// configured (svc.knowledgeBase nil).
func mountKnowledgeBaseService(mux *http.ServeMux, svc *serviceContainer, opts []connect.HandlerOption) {
	if svc.knowledgeBase == nil {
		return
	}
	srv := knowledgebaseconnect.NewServer(svc.knowledgeBase, svc.org, svc.kbSyncWorker)
	knowledgebaseconnect.Mount(mux, srv, opts...)
}

// mountWorkflowService wires reusable Workflow definitions and their runs.
func mountWorkflowService(mux *http.ServeMux, svc *serviceContainer, rest *v1.Services, opts []connect.HandlerOption) {
	if svc.workflow == nil || rest == nil || rest.WorkflowOrchestrator == nil {
		return
	}
	var podTerm workflowconnect.PodTerminatorForWorkflow
	if rest.PodCoordinator != nil {
		podTerm = rest.PodCoordinator
	}
	srv := workflowconnect.NewServer(svc.workflow, svc.workflowRun, rest.WorkflowOrchestrator, svc.org, podTerm)
	workflowconnect.Mount(mux, srv, opts...)
}

// mountLicenseService wires both LicenseService (auth-required: Activate /
// Refresh / Validate) and LicensePublicService (no auth: GetStatus /
// GetLimits / CheckFeature) onto the mux. The public service intentionally
// mounts WITHOUT `opts` — the auth interceptor would reject the unauthenticated
// status check the login page hits before any token exists. Conventions §3.5
// exception #1 (system-wide config, not org-scoped).
//
// Skips when the license service is nil — non-OnPremise deployments don't
// initialize the subsystem (services_init_helpers.go:initializeLicenseService),
// so neither the auth-required nor the public surface should advertise itself.
func mountLicenseService(mux *http.ServeMux, svc *serviceContainer, opts []connect.HandlerOption) {
	if svc.license == nil {
		return
	}
	licenseconnect.Mount(mux, licenseconnect.NewServer(svc.license), opts...)
	licenseconnect.MountPublic(mux, licenseconnect.NewPublicServer(svc.license))
}
