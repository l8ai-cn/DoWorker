package agentpod

import (
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
)

// OrchestrateCreatePodRequest is the unified Pod creation request (protocol-agnostic).
// Pod configuration flows exclusively through AgentfileLayer (SSOT).
type OrchestrateCreatePodRequest struct {
	OrganizationID int64
	UserID         int64

	RunnerID       int64
	AgentSlug      string
	RepositoryID   *int64 // Platform-level ID (from AgentFile REPO slug resolution or resume inheritance)
	TicketID       *int64
	TicketSlug     *string
	Alias          *string
	AgentfileLayer *string // SSOT for all CONFIG, MODE, PROMPT, REPO, BRANCH, USE_ENV_BUNDLE
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

	// SessionConfigBundles are ephemeral config-kind bundles (name → parsed
	// JSON doc) resolved per-session, consumed by USE_CONFIG_BUNDLE in the
	// AgentfileLayer. Used by the model-pool flow to inject the do-agent
	// settings.json (provider key + model) without persisting a bundle row.
	SessionConfigBundles map[string]interface{}

	// SessionEnvBundles are ephemeral credential env maps (name → KV) consumed
	// by USE_ENV_BUNDLE. The model-pool flow injects codex OPENAI_* here.
	SessionEnvBundles map[string]map[string]string

	// Worker model binding (quota/billing). VirtualAPIKeyID binds a virtual
	// key (its wrapped ai_models credential is injected, usage attributed to
	// the key). ModelConfigID binds a real ai_models row directly. TokenBudget
	// is an informational per-Worker hint surfaced to the harness.
	ModelConfigID   *int64
	VirtualAPIKeyID *int64
	TokenBudget     *int64

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
