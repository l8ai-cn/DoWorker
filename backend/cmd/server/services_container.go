package main

import (
	agentpodDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/extension"
	runnerDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/database"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/email"
	adminservice "github.com/l8ai-cn/agentcloud/backend/internal/service/admin"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agent"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	agentsessionsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentsession"
	airesourceservice "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
	apikeyservice "github.com/l8ai-cn/agentcloud/backend/internal/service/apikey"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/auth"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/billing"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/binding"
	blockstoreservice "github.com/l8ai-cn/agentcloud/backend/internal/service/blockstore"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/channel"
	envbundleservice "github.com/l8ai-cn/agentcloud/backend/internal/service/envbundle"
	executionclusterservice "github.com/l8ai-cn/agentcloud/backend/internal/service/executioncluster"
	extensionservice "github.com/l8ai-cn/agentcloud/backend/internal/service/extension"
	fileservice "github.com/l8ai-cn/agentcloud/backend/internal/service/file"
	goalloop "github.com/l8ai-cn/agentcloud/backend/internal/service/goalloop"
	grantservice "github.com/l8ai-cn/agentcloud/backend/internal/service/grant"
	imbridgesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/imbridge"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/invitation"
	knowledgebaseservice "github.com/l8ai-cn/agentcloud/backend/internal/service/knowledgebase"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/license"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/mesh"
	notifservice "github.com/l8ai-cn/agentcloud/backend/internal/service/notification"
	orchestrationcontrol "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	orchestrationworker "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationworker"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/organization"
	permissionpolicysvc "github.com/l8ai-cn/agentcloud/backend/internal/service/permissionpolicy"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/promocode"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/repository"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	podsessionsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/sessionusage"
	ssoservice "github.com/l8ai-cn/agentcloud/backend/internal/service/sso"
	supportticketservice "github.com/l8ai-cn/agentcloud/backend/internal/service/supportticket"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/ticket"
	tokenquotasvc "github.com/l8ai-cn/agentcloud/backend/internal/service/tokenquota"
	tokenusagesvc "github.com/l8ai-cn/agentcloud/backend/internal/service/tokenusage"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/user"
	virtualkeysvc "github.com/l8ai-cn/agentcloud/backend/internal/service/virtualkey"
	workflow "github.com/l8ai-cn/agentcloud/backend/internal/service/workflow"
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
	agentSessions       *agentsessionsvc.Service
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
