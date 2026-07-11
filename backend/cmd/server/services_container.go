package main

import (
	agentpodDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/domain/extension"
	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/anthropics/agentsmesh/backend/internal/infra/database"
	"github.com/anthropics/agentsmesh/backend/internal/infra/email"
	adminservice "github.com/anthropics/agentsmesh/backend/internal/service/admin"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	airesourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	apikeyservice "github.com/anthropics/agentsmesh/backend/internal/service/apikey"
	"github.com/anthropics/agentsmesh/backend/internal/service/auth"
	"github.com/anthropics/agentsmesh/backend/internal/service/billing"
	"github.com/anthropics/agentsmesh/backend/internal/service/binding"
	blockstoreservice "github.com/anthropics/agentsmesh/backend/internal/service/blockstore"
	"github.com/anthropics/agentsmesh/backend/internal/service/channel"
	envbundleservice "github.com/anthropics/agentsmesh/backend/internal/service/envbundle"
	extensionservice "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	fileservice "github.com/anthropics/agentsmesh/backend/internal/service/file"
	grantservice "github.com/anthropics/agentsmesh/backend/internal/service/grant"
	imbridgesvc "github.com/anthropics/agentsmesh/backend/internal/service/imbridge"
	"github.com/anthropics/agentsmesh/backend/internal/service/invitation"
	knowledgebaseservice "github.com/anthropics/agentsmesh/backend/internal/service/knowledgebase"
	"github.com/anthropics/agentsmesh/backend/internal/service/license"
	loop "github.com/anthropics/agentsmesh/backend/internal/service/loop"
	"github.com/anthropics/agentsmesh/backend/internal/service/mesh"
	notifservice "github.com/anthropics/agentsmesh/backend/internal/service/notification"
	"github.com/anthropics/agentsmesh/backend/internal/service/organization"
	permissionpolicysvc "github.com/anthropics/agentsmesh/backend/internal/service/permissionpolicy"
	"github.com/anthropics/agentsmesh/backend/internal/service/promocode"
	"github.com/anthropics/agentsmesh/backend/internal/service/repository"
	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	podsessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionusage"
	ssoservice "github.com/anthropics/agentsmesh/backend/internal/service/sso"
	supportticketservice "github.com/anthropics/agentsmesh/backend/internal/service/supportticket"
	"github.com/anthropics/agentsmesh/backend/internal/service/ticket"
	tokenquotasvc "github.com/anthropics/agentsmesh/backend/internal/service/tokenquota"
	tokenusagesvc "github.com/anthropics/agentsmesh/backend/internal/service/tokenusage"
	"github.com/anthropics/agentsmesh/backend/internal/service/user"
	virtualkeysvc "github.com/anthropics/agentsmesh/backend/internal/service/virtualkey"
)

var _ runner.PodStore = (*agentpod.PodService)(nil)

type serviceContainer struct {
	auth               *auth.Service
	user               *user.Service
	org                *organization.Service
	admin              *adminservice.Service
	adminDB            database.DB
	agentSvc           *agent.AgentService
	envBundle          *envbundleservice.Service
	userConfig         *agent.UserConfigService
	repository         *repository.Service
	webhook            *repository.WebhookService
	runner             *runner.Service
	pod                *agentpod.PodService
	autopilot          *agentpod.AutopilotControllerService
	channel            *channel.Service
	ticket             *ticket.Service
	mrSync             *ticket.MRSyncService
	billing            *billing.Service
	binding            *binding.Service
	mesh               *mesh.Service
	message            *agent.MessageService
	invitation         *invitation.Service
	file               *fileservice.Service
	promoCode          *promocode.Service
	agentpodSettings   *agentpod.SettingsService
	agentpodAIProvider *agentpod.AIProviderService
	aiResource         *airesourceservice.Service
	virtualKey         *virtualkeysvc.Service
	tokenQuota         *tokenquotasvc.Service
	license            *license.Service
	apikey             *apikeyservice.Service
	apikeyAdapter      *apikeyservice.MiddlewareAdapter
	email              email.Service
	extension          *extensionservice.Service
	extensionRepo      extension.Repository
	marketplaceWorker  *extensionservice.MarketplaceWorker
	loop               *loop.LoopService
	loopRun            *loop.LoopRunService
	sso                *ssoservice.Service
	supportTicket      *supportticketservice.Service
	tokenUsage         *tokenusagesvc.Service
	podSessionUsage    *podsessionsvc.Service
	permissionPolicy   *permissionpolicysvc.Service
	blockstore         *blockstoreservice.Service
	grant              *grantservice.Service
	knowledgeBase      *knowledgebaseservice.Service
	kbSyncWorker       *knowledgebaseservice.SyncWorker
	workerServices
	imBridge        *imbridgesvc.Bridge
	notifDispatcher *notifservice.Dispatcher
	notifPrefStore  *notifservice.PreferenceStore
	podRepo         agentpodDomain.PodRepository
	runnerRepo      runnerDomain.RunnerRepository
	autopilotRepo   agentpodDomain.AutopilotRepository
}

func (s *serviceContainer) Close() {
	if s != nil && s.blockstore != nil {
		s.blockstore.Close()
	}
}
