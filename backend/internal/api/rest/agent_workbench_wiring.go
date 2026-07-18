package rest

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/api/rest/v1"
	agentworkbenchdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	agentworkbenchsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentworkbench"
	sessionfilesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionfile"
	sessionmessagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionmessage"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func wireAgentWorkbench(
	services *v1.Services,
	db *gorm.DB,
	sessions *sessionsvc.Service,
) {
	repository := infra.NewAgentWorkbenchRepository(db)
	hub := agentworkbenchsvc.NewDeltaHub(128)
	sessionFiles := sessionfilesvc.NewService(db, services.File)
	ingress, err := agentworkbenchsvc.NewIngress(
		sessions,
		services.Pod,
		repository,
		hub,
		agentworkbenchsvc.NewSessionFileArtifactMaterializer(
			sessionFiles,
			services.SandboxFsService,
		),
		uuid.NewString,
	)
	if err != nil {
		panic(fmt.Sprintf("initialize agent workbench ingress: %v", err))
	}
	services.AgentWorkbenchRepo = repository
	services.AgentWorkbenchHub = hub
	services.AgentWorkbenchIngress = ingress
	services.AgentSessions = sessions
	wireAgentWorkbenchCommands(services, db, repository, hub)
	if services.RunnerGRPCAdapter != nil {
		services.RunnerGRPCAdapter.SetWorkbenchEventSink(ingress)
	}
}

func wireAgentWorkbenchCommands(
	services *v1.Services,
	db *gorm.DB,
	repository agentworkbenchdomain.Repository,
	hub *agentworkbenchsvc.DeltaHub,
) {
	if services.Pod == nil || services.PendingQueue == nil ||
		services.PodCoordinator == nil {
		slog.Warn("agent workbench commands unavailable: runner queue is not wired")
		return
	}
	dispatcher, err := agentworkbenchsvc.NewCommandDispatcher(
		repository,
		services.Pod,
		sessionmessagesvc.NewPromptOutbox(db, services.PendingQueue),
		services.PodCoordinator.GetCommandSender(),
		hub,
		time.Now,
		agentworkbenchsvc.WithAttachmentDelivery(
			sessionfilesvc.NewService(db, services.File),
			services.SandboxFsService,
		),
	)
	if err != nil {
		panic(fmt.Sprintf("initialize agent workbench commands: %v", err))
	}
	services.AgentWorkbenchCommands = dispatcher
}
