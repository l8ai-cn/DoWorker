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
	agentsessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	airesourceservice "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	apikeyservice "github.com/anthropics/agentsmesh/backend/internal/service/apikey"
	"github.com/anthropics/agentsmesh/backend/internal/service/auth"
	"github.com/anthropics/agentsmesh/backend/internal/service/billing"
	"github.com/anthropics/agentsmesh/backend/internal/service/binding"
	blockstoreservice "github.com/anthropics/agentsmesh/backend/internal/service/blockstore"
	"github.com/anthropics/agentsmesh/backend/internal/service/channel"
	conversationitemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	envbundleservice "github.com/anthropics/agentsmesh/backend/internal/service/envbundle"
	executionclusterservice "github.com/anthropics/agentsmesh/backend/internal/service/executioncluster"
	extensionservice "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	fileservice "github.com/anthropics/agentsmesh/backend/internal/service/file"
	goalloop "github.com/anthropics/agentsmesh/backend/internal/service/goalloop"
	grantservice "github.com/anthropics/agentsmesh/backend/internal/service/grant"
	imbridgesvc "github.com/anthropics/agentsmesh/backend/internal/service/imbridge"
	"github.com/anthropics/agentsmesh/backend/internal/service/invitation"
	knowledgebaseservice "github.com/anthropics/agentsmesh/backend/internal/service/knowledgebase"
	"github.com/anthropics/agentsmesh/backend/internal/service/license"
	"github.com/anthropics/agentsmesh/backend/internal/service/mesh"
	notifservice "github.com/anthropics/agentsmesh/backend/internal/service/notification"
	orchestrationcontrol "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	orchestrationworker "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"
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
	workflow "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
)

var _ runner.PodStore = (*agentpod.PodService)(nil)

type serviceContainer struct {
	auth                *auth.Service
	user                *user.Service
	org                 *organization.Service
	admin               *adminservice.Service
	adminDB             database.DB
	agentSvc            *agent.AgentService
	envBundle           *envbundleservice.Service
	userConfig          *agent.UserConfigService
	repository          *repository.Service
	webhook             *repository.WebhookService
	runner              *runner.Service
	executionCluster    *executionclusterservice.Service
	pod                 *agentpod.PodService
	autopilot           *agentpod.AutopilotControllerService
	channel             *channel.Service
	ticket              *ticket.Service
	mrSync              *ticket.MRSyncService
	billing             *billing.Service
	binding             *binding.Service
	mesh                *mesh.Service
	message             *agent.MessageService
	invitation          *invitation.Service
	file                *fileservice.Service
	promoCode           *promocode.Service
	agentpodSettings    *agentpod.SettingsService
	agentpodAIProvider  *agentpod.AIProviderService
	aiResource          *airesourceservice.Service
	virtualKey          *virtualkeysvc.Service
	tokenQuota          *tokenquotasvc.Service
	license             *license.Service
	apikey              *apikeyservice.Service
	apikeyAdapter       *apikeyservice.MiddlewareAdapter
	email               email.Service
	extension           *extensionservice.Service
	extensionRepo       extension.Repository
	marketplaceWorker   *extensionservice.MarketplaceWorker
	workflow            *workflow.WorkflowService
	workflowRun         *workflow.WorkflowRunService
	goalLoop            *goalloop.Service
	sso                 *ssoservice.Service
	supportTicket       *supportticketservice.Service
	tokenUsage          *tokenusagesvc.Service
	podSessionUsage     *podsessionsvc.Service
	permissionPolicy    *permissionpolicysvc.Service
	blockstore          *blockstoreservice.Service
	grant               *grantservice.Service
	knowledgeBase       *knowledgebaseservice.Service
	kbSyncWorker        *knowledgebaseservice.SyncWorker
	orchestration       *orchestrationcontrol.Service
	bindingApply        *orchestrationworker.BindingApplyService
	workerTemplateApply *orchestrationworker.WorkerTemplateApplyService
	promptApply         *orchestrationworker.PromptApplyService
	expertApply         *orchestrationworker.ExpertApplyService
	workflowApply       *orchestrationworker.WorkflowApplyService
	goalLoopApply       *orchestrationworker.GoalLoopApplyService
	workerApply         *orchestrationworker.WorkerApplyService
	workerApplyRuntime  orchestrationWorkerApplyRuntime
	workerServices
	imBridge *imbridgesvc.Bridge

	notifDispatcher *notifservice.Dispatcher
	notifPrefStore  *notifservice.PreferenceStore

	podRepo       agentpodDomain.PodRepository
	runnerRepo    runnerDomain.RunnerRepository
	autopilotRepo agentpodDomain.AutopilotRepository
}

func (s *serviceContainer) Close() {
	if s != nil && s.blockstore != nil {
		s.blockstore.Close()
	}
}
