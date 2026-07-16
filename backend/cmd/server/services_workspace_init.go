package main

import (
	"log/slog"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/infra/email"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/service/billing"
	"github.com/anthropics/agentsmesh/backend/internal/service/binding"
	"github.com/anthropics/agentsmesh/backend/internal/service/channel"
	executionclusterservice "github.com/anthropics/agentsmesh/backend/internal/service/executioncluster"
	grantservice "github.com/anthropics/agentsmesh/backend/internal/service/grant"
	imbridgesvc "github.com/anthropics/agentsmesh/backend/internal/service/imbridge"
	"github.com/anthropics/agentsmesh/backend/internal/service/invitation"
	"github.com/anthropics/agentsmesh/backend/internal/service/mesh"
	"github.com/anthropics/agentsmesh/backend/internal/service/organization"
	"github.com/anthropics/agentsmesh/backend/internal/service/promocode"
	"github.com/anthropics/agentsmesh/backend/internal/service/repository"
	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"github.com/anthropics/agentsmesh/backend/internal/service/ticket"
	tokenquotasvc "github.com/anthropics/agentsmesh/backend/internal/service/tokenquota"
	virtualkeysvc "github.com/anthropics/agentsmesh/backend/internal/service/virtualkey"
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
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
		cfg, db, services.agentSvc, services.aiResource, services.repository, services.runner,
	)
	if err != nil {
		return err
	}
	services.workerServices = workerServices
	return attachOrchestrationControl(services, db)
}
