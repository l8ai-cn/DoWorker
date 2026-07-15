package agentpod

import (
	"context"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessionDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

// OrchestrateCreatePodRequest is the unified Pod creation request (protocol-agnostic).
// Pod configuration flows exclusively through AgentfileLayer (SSOT).
type OrchestrateCreatePodRequest struct {
	OrganizationID int64
	UserID         int64

	RunnerID                 int64
	AgentSlug                string
	RepositoryID             *int64 // Platform-level ID (from AgentFile REPO slug resolution or resume inheritance)
	TicketID                 *int64
	TicketSlug               *string
	Alias                    *string
	AgentfileLayer           *string // SSOT for all CONFIG, MODE, PROMPT, REPO, BRANCH, USE_ENV_BUNDLE
	WorkerSpecDraft          *workercreation.Draft
	WorkerSpecSnapshotID     *int64
	WorkerSpecPromptOverride *string
	// AutomationLevel is the unified permission/automation tier requested at
	// creation (interactive/auto_edit/autonomous). Empty ⇒ autonomous default.
	// The orchestrator translates it into agent-native CONFIG/MODE layer lines.
	AutomationLevel string
	Cols            int32
	Rows            int32

	SourcePodKey            string
	ResumeAgentSession      *bool
	ResumeExternalSessionID string
	AgentSessionID          string
	SessionMcpServers       map[string]interface{}
	SessionProvision        *sessionDomain.ProvisionSpec
	PrepareSession          func(context.Context, *sessionDomain.Session) error

	// SessionConfigBundles are ephemeral config-kind documents consumed by
	// USE_CONFIG_BUNDLE declarations.
	SessionConfigBundles map[string]interface{}

	ModelResourceID   *int64
	ModelResourceEnv  map[string]string
	ModelResourceArgs []string
	TokenBudget       *int64

	Perpetual           bool
	DeferRunnerDispatch bool
	BranchName          *string
	// KnowledgeMounts are per-pod KB selections; they win over Agentfile
	// KNOWLEDGE declarations and agent default mounts on mode conflicts.
	KnowledgeMounts []KnowledgeMountRequest
	// LocalPath is an absolute directory on the runner host (Omnigent compat
	// workspace picker). Maps to SandboxConfig.local_path — not agentfile syntax.
	LocalPath          string
	QueueIfUnavailable bool
	QueueTTL           time.Duration

	preResolvedRepository     *gitprovider.Repository
	preResolvedRepositorySlug string
	resolvedWorkerSpec        *specservice.ResolvedSnapshot
	preparedWorkerSpec        *specdomain.Spec
	workerSpecSnapshotID      *int64
}

// KnowledgeMountRequest selects one knowledge base for the pod being created.
type KnowledgeMountRequest struct {
	Slug string
	Mode string // ro | rw; empty defaults to ro
}

type OrchestrateCreatePodResult struct {
	Pod     *podDomain.Pod
	Warning string
	Queued  bool
}
