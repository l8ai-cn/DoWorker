package rest

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/api/rest/v1"
	sessionapi "github.com/anthropics/agentsmesh/backend/internal/api/rest/v1/session"
	"github.com/anthropics/agentsmesh/backend/internal/config"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	permissionpolicysvc "github.com/anthropics/agentsmesh/backend/internal/service/permissionpolicy"
	commentsvc "github.com/anthropics/agentsmesh/backend/internal/service/sessioncomment"
	sessionfilesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionfile"
	sessionmessagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionmessage"
	permgrantsvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionpermission"
	sessionusagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionusage"
	tokenquotasvc "github.com/anthropics/agentsmesh/backend/internal/service/tokenquota"
	"github.com/anthropics/agentsmesh/backend/pkg/embedtoken"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func registerSessionRoutes(
	r *gin.Engine,
	cfg *config.Config,
	svc *v1.Services,
	db *gorm.DB,
	redisClient *redis.Client,
) {
	sessions := svc.AgentSessions
	if sessions == nil {
		sessions = sessionsvc.NewService(db)
	}
	sessionDeps := sessionapi.Deps{
		Auth:               svc.Auth,
		User:               svc.User,
		Org:                svc.Org,
		Agent:              svc.AgentSvc,
		Runner:             svc.Runner,
		Sessions:           sessions,
		Items:              itemsvc.NewService(db),
		Hub:                sessionapi.NewSessionHub(),
		Elicitations:       sessionapi.NewElicitationStore(),
		PodOrchestrator:    svc.PodOrchestrator,
		WorkerCreation:     svc.WorkerCreation,
		Pod:                svc.Pod,
		DeferredCommitter:  sessionsvc.NewDeferredCommitter(db),
		DispatchQueue:      svc.PendingQueue,
		RelayManager:       svc.RelayManager,
		RelayTokens:        svc.RelayTokenGenerator,
		SessionUsage:       sessionusagesvc.NewService(db),
		Policies:           permissionpolicysvc.NewService(db),
		ReadState:          sessionapi.NewReadStateStore(db),
		SandboxFs:          svc.SandboxFsService,
		SessionFiles:       sessionfilesvc.NewService(db, svc.File),
		MessageOutbox:      sessionmessagesvc.NewPromptOutbox(db, svc.PendingQueue),
		SessionComments:    commentsvc.NewService(db),
		SessionPermissions: permgrantsvc.NewService(db),
		Grants:             svc.Grant,
		AIResources:        svc.AIResource,
		EnvBundles:         svc.EnvBundle,
		VirtualKeys:        svc.VirtualKey,
		TokenQuotas:        tokenquotasvc.NewService(infra.NewTokenQuotaRepository(db), db),
		EmbedTokens:        embedtoken.NewService(cfg.JWT.Secret, redisClient),
		Version:            "do-worker-dev",
	}
	sessionDeps.Stream = sessionapi.NewSessionStreamPublisher(
		sessionDeps.Hub, sessionDeps.Items, sessionDeps.Sessions, sessionDeps.Elicitations,
	)
	sessionDeps.Stream.Usage = sessionDeps.SessionUsage
	sessionDeps.Stream.Pods = svc.Pod
	sessionDeps.Updates = sessionapi.NewSessionUpdatesHub(&sessionDeps)
	sessionDeps.Stream.Updates = sessionDeps.Updates
	svc.EmbedTokens = sessionDeps.EmbedTokens
	wireAgentWorkbench(svc, db, sessions)
	sessionDeps.WorkbenchRepo = svc.AgentWorkbenchRepo
	if svc.PodCoordinator != nil {
		sessionDeps.PodCoordinator = svc.PodCoordinator
		sessionDeps.CommandSender = svc.PodCoordinator.GetCommandSender()
		podEvents := sessionDeps.Stream
		svc.PodCoordinator.AddStatusChangeCallback(func(podKey, podStatus, agentStatus string) {
			podEvents.PublishPodStatus(context.Background(), podKey, podStatus, agentStatus)
		})
	}
	sessionapi.RegisterHealthRoute(r, sessionDeps)
	sessionapi.RegisterRoutes(r, sessionDeps)
	if svc.RunnerGRPCAdapter != nil {
		svc.RunnerGRPCAdapter.SetPodEventSink(sessionDeps.Stream)
	}
}
