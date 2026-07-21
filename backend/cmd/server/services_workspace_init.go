package main

import (
	"log/slog"

	"github.com/l8ai-cn/agentcloud/backend/internal/config"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/email"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agent"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	agentsessionsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentsession"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/billing"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/binding"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/channel"
	executionclusterservice "github.com/l8ai-cn/agentcloud/backend/internal/service/executioncluster"
	grantservice "github.com/l8ai-cn/agentcloud/backend/internal/service/grant"
	imbridgesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/imbridge"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/invitation"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/mesh"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/organization"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/promocode"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/repository"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/ticket"
	tokenquotasvc "github.com/l8ai-cn/agentcloud/backend/internal/service/tokenquota"
	virtualkeysvc "github.com/l8ai-cn/agentcloud/backend/internal/service/virtualkey"
	"github.com/l8ai-cn/agentcloud/backend/pkg/crypto"
	"gorm.io/gorm"
)

func initializeWorkspaceServices(services *serviceContainer, cfg *config.Config, db *gorm.DB, encryptor *crypto.Encryptor) error {
	gitRepoRepo := infra.NewGitProviderRepository(db)
	services.repository = repository.NewService(gitRepoRepo)
	services.webhook = repository.NewWebhookService(gitRepoRepo, cfg, services.user, slog.Default())
	services.repository.SetWebhookService(services.webhook)

	services.billing = billing.NewServiceWithConfig(infra.NewBillingRepository(db), cfg)
	services.org = organization.NewServiceWithBilling(infra.NewOrganizationRepository(db), services.billing)
	services.aiResource = initializeAIResourceService(db, services.org, encryptor)
	services.runnerRepo = infra.NewRunnerRepository(db)
	services.runner = runner.NewService(services.runnerRepo, services.billing)
	clusterRepo := infra.NewExecutionClusterRepository(db)
	services.runner.SetExecutionClusterRepository(clusterRepo)
	services.executionCluster = executionclusterservice.NewService(clusterRepo, services.runnerRepo, services.runner, cfg.BaseURL())
	grantRepo := infra.NewGrantRepository(db)
	services.grant = grantservice.NewService(grantRepo)
	services.runner.SetGrantQuerier(grantRepo)

	services.podRepo = infra.NewPodRepository(db)
	services.pod = agentpod.NewPodService(services.podRepo)
	services.agentSessions = agentsessionsvc.NewService(db)
	services.autopilotRepo = infra.NewAutopilotRepository(db)
	services.autopilot = agentpod.NewAutopilotControllerService(services.autopilotRepo)
	services.channel = channel.NewService(infra.NewChannelRepository(db))
	imBridgeSvc := imbridgesvc.NewService(infra.NewIMBridgeRepository(db), imbridgesvc.NewRegistry(nil), cfg.BaseURL())
	services.imBridge = imbridgesvc.NewBridge(imBridgeSvc, services.channel)
	services.ticket = ticket.NewService(infra.NewTicketRepository(db))
	services.mrSync = ticket.NewMRSyncService(infra.NewMRSyncRepository(db), nil)
	services.binding = binding.NewService(infra.NewBindingRepository(db), services.pod)
	services.mesh = mesh.NewService(infra.NewMeshRepository(db), services.pod, services.channel, services.binding)
	services.message = agent.NewMessageService(infra.NewAgentMessageRepository(db))

	services.email = email.NewService(email.Config{
		Provider: cfg.Email.Provider, ResendKey: cfg.Email.ResendKey,
		FromAddress: cfg.Email.FromAddress, BaseURL: cfg.FrontendURL(),
	})
	services.invitation = invitation.NewService(infra.NewInvitationRepository(db), services.email)
	services.promoCode = promocode.NewService(infra.NewPromocodeRepository(db), infra.NewGormBillingProvider(db))
	services.agentpodSettings = agentpod.NewSettingsService(infra.NewSettingsRepository(db))
	services.agentpodAIProvider = agentpod.NewAIProviderService(infra.NewAIProviderRepository(db), encryptor)
	services.virtualKey = virtualkeysvc.NewService(infra.NewVirtualAPIKeyRepository(db), services.aiResource)
	services.tokenQuota = tokenquotasvc.NewService(infra.NewTokenQuotaRepository(db), db)
	workerServices, err := initializeWorkerServices(
		cfg, db, services.agentSvc, services.aiResource, services.repository, services.runner, services.user,
	)
	if err != nil {
		return err
	}
	services.workerServices = workerServices
	return attachOrchestrationControl(services, db)
}
