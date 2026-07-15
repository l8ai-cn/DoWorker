package agentpod

import (
	"context"
	"testing"

	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/anthropics/agentsmesh/backend/internal/domain/ticket"
	"github.com/anthropics/agentsmesh/backend/internal/domain/user"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	userService "github.com/anthropics/agentsmesh/backend/internal/service/user"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"gorm.io/gorm"
)

// ==================== Mock Definitions ====================

// mockPodCoordinator implements PodCoordinatorForOrchestrator.
type mockPodCoordinator struct {
	createPodCalled bool
	lastRunnerID    int64
	lastCmd         *runnerv1.CreatePodCommand
	lastQueueOpts   podDomain.CreatePodQueueOpts
	queueErr        error
	err             error
}

func (m *mockPodCoordinator) CreatePod(_ context.Context, runnerID int64, cmd *runnerv1.CreatePodCommand) error {
	m.createPodCalled = true
	m.lastRunnerID = runnerID
	m.lastCmd = cmd
	return m.err
}

func (m *mockPodCoordinator) CreatePodOrQueue(ctx context.Context, runnerID int64, cmd *runnerv1.CreatePodCommand, opts podDomain.CreatePodQueueOpts) error {
	m.lastQueueOpts = opts
	if m.queueErr != nil {
		m.lastRunnerID = runnerID
		m.lastCmd = cmd
		return m.queueErr
	}
	return m.CreatePod(ctx, runnerID, cmd)
}

// mockBillingService implements BillingServiceForOrchestrator.
type mockBillingService struct {
	err error
}

func (m *mockBillingService) CheckQuota(_ context.Context, _ int64, _ string, _ int) error {
	return m.err
}

// mockUserServiceForOrch implements UserServiceForOrchestrator.
type mockUserServiceForOrch struct {
	defaultCred    *user.GitCredential
	defaultCredErr error
	decryptedCred  *userService.DecryptedCredential
	decryptedErr   error
}

func (m *mockUserServiceForOrch) GetDefaultGitCredential(_ context.Context, _ int64) (*user.GitCredential, error) {
	return m.defaultCred, m.defaultCredErr
}

func (m *mockUserServiceForOrch) GetDecryptedCredentialToken(_ context.Context, _, _ int64) (*userService.DecryptedCredential, error) {
	return m.decryptedCred, m.decryptedErr
}

// mockRepoService implements RepositoryServiceForOrchestrator.
type mockRepoService struct {
	repo                *gitprovider.Repository
	err                 error
	getAccessibleCalls  []repositoryAccessCall
	findAccessibleCalls []repositorySlugAccessCall
}

type repositoryAccessCall struct {
	ID             int64
	OrganizationID int64
	UserID         int64
}

type repositorySlugAccessCall struct {
	OrganizationID int64
	UserID         int64
	Slug           string
}

func (m *mockRepoService) GetAccessibleByID(_ context.Context, id, orgID, userID int64) (*gitprovider.Repository, error) {
	m.getAccessibleCalls = append(m.getAccessibleCalls, repositoryAccessCall{
		ID:             id,
		OrganizationID: orgID,
		UserID:         userID,
	})
	return m.repo, m.err
}

func (m *mockRepoService) FindAccessibleByOrgSlug(_ context.Context, orgID, userID int64, slug string) (*gitprovider.Repository, error) {
	m.findAccessibleCalls = append(m.findAccessibleCalls, repositorySlugAccessCall{
		OrganizationID: orgID,
		UserID:         userID,
		Slug:           slug,
	})
	return m.repo, m.err
}

// mockTicketServiceForOrch implements TicketServiceForOrchestrator.
type mockTicketServiceForOrch struct {
	ticket *ticket.Ticket
	err    error
}

func (m *mockTicketServiceForOrch) GetTicket(_ context.Context, _ int64) (*ticket.Ticket, error) {
	return m.ticket, m.err
}

func (m *mockTicketServiceForOrch) GetTicketBySlug(_ context.Context, _ int64, _ string) (*ticket.Ticket, error) {
	return m.ticket, m.err
}

// mockAgentConfigProvider implements agent.AgentConfigProvider for ConfigBuilder.
// After the EnvBundle refactor only GetAgent is required from the provider —
// credential resolution now lives in ConfigBuilder.envBundleSvc. The legacy
// `creds`/`isRunner` fields are retained for tests that still want to assert
// credential injection, but they're consumed via the mockEnvBundleProvider
// wired into ConfigBuilder, not via this interface.
type mockAgentConfigProvider struct {
	agentDef *agentDomain.Agent
	agentErr error
	config   agentDomain.ConfigValues
	creds    agentDomain.EncryptedCredentials
	isRunner bool
	credsErr error
}

func (m *mockAgentConfigProvider) GetAgent(_ context.Context, _ string) (*agentDomain.Agent, error) {
	return m.agentDef, m.agentErr
}

// mockRunnerSelector implements RunnerSelectorForOrchestrator for testing.
type mockRunnerSelector struct {
	runner        *runnerDomain.Runner
	err           error
	selectCalled  bool
	selectHints   *runnerDomain.AffinityHints
	resolveRunner *runnerDomain.Runner
	resolveErr    error
	resolveCall   *runnerResolveCall
}

func (m *mockRunnerSelector) SelectRunnerWithAffinity(_ context.Context, _ int64, _ int64, _ string, hints *runnerDomain.AffinityHints, _ map[int64]int) (*runnerDomain.Runner, error) {
	m.selectCalled = true
	m.selectHints = hints
	return m.runner, m.err
}

type runnerResolveCall struct {
	RunnerID         int64
	OrganizationID   int64
	UserID           int64
	AgentSlug        string
	AllowUnavailable bool
}

func (m *mockRunnerSelector) ResolveRunnerForCreate(
	_ context.Context,
	runnerID, orgID, userID int64,
	agentSlug string,
	allowUnavailable bool,
) (*runnerDomain.Runner, error) {
	m.resolveCall = &runnerResolveCall{
		RunnerID:         runnerID,
		OrganizationID:   orgID,
		UserID:           userID,
		AgentSlug:        agentSlug,
		AllowUnavailable: allowUnavailable,
	}
	return m.resolveRunner, m.resolveErr
}

// mockAgentResolver implements AgentResolverForOrchestrator for testing.
type mockAgentResolver struct {
	agentDef *agentDomain.Agent
	err      error
	calls    int
}

func (m *mockAgentResolver) GetAgent(_ context.Context, _ string) (*agentDomain.Agent, error) {
	m.calls++
	return m.agentDef, m.err
}

// ==================== Helper Functions ====================

// setupOrchestratorTestDB extends setupTestDB with additional tables required
// by GORM Preload in GetPod (agents, repositories).
// We keep setupTestDB unchanged to avoid breaking existing tests.
func setupOrchestratorTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)

	// agents table — needed by Preload("Agent") when AgentSlug is set
	db.Exec(`CREATE TABLE IF NOT EXISTS agents (
		id INTEGER PRIMARY KEY,
		slug TEXT,
		name TEXT,
		launch_command TEXT,
		adapter_id TEXT NOT NULL DEFAULT '',
		description TEXT,
		config_schema TEXT DEFAULT '{}',
		agentfile_source TEXT,
		supported_modes TEXT NOT NULL DEFAULT 'pty',
		uses_legacy_columns INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)

	// repositories table — needed by Preload("Repository") when RepositoryID is set
	db.Exec(`CREATE TABLE IF NOT EXISTS repositories (
		id INTEGER PRIMARY KEY,
		organization_id INTEGER,
		provider_type TEXT,
		provider_base_url TEXT,
		clone_url TEXT,
		http_clone_url TEXT,
		ssh_clone_url TEXT,
		external_id TEXT,
		name TEXT,
		slug TEXT,
		default_branch TEXT DEFAULT 'main',
		preparation_script TEXT,
		preparation_timeout INTEGER DEFAULT 300,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)

	return db
}

func newTestProvider() *mockAgentConfigProvider {
	agentfile := `
AGENT claude
EXECUTABLE claude
MCP ON
arg "--session-id" config.session_id when config.session_id != "" and not config.resume_enabled
arg "--resume" config.resume_session when config.resume_enabled
PROMPT_POSITION prepend
`
	return &mockAgentConfigProvider{
		agentDef: &agentDomain.Agent{
			Slug:              "claude-code",
			Name:              "Claude Code",
			LaunchCommand:     "claude",
			AdapterID:         "claude-stream-json",
			SupportedModes:    "pty",
			AgentfileSource:   &agentfile,
			UsesLegacyColumns: true,
		},
		config:   agentDomain.ConfigValues{},
		creds:    agentDomain.EncryptedCredentials{},
		isRunner: true,
	}
}

func newCodexTestProvider() *mockAgentConfigProvider {
	agentfile := `
AGENT codex
EXECUTABLE codex
MODE pty
MODE acp "app-server"
CONFIG approval_mode SELECT("untrusted", "on-request", "never") = "untrusted"
ENV OPENAI_API_KEY SECRET OPTIONAL
ENV CODEX_HOME = sandbox.root + "/codex-home"
PROMPT_POSITION append
MCP ON
arg "resume" "--last" when config.resume_enabled and mode != "acp"
arg "--ask-for-approval" config.approval_mode when config.approval_mode != "" and mode != "acp"
`
	return &mockAgentConfigProvider{
		agentDef: &agentDomain.Agent{
			Slug:            "codex-cli",
			Name:            "Codex CLI",
			LaunchCommand:   "codex",
			AdapterID:       "codex-app-server",
			SupportedModes:  "pty,acp",
			AgentfileSource: &agentfile,
		},
		config:   agentDomain.ConfigValues{},
		creds:    agentDomain.EncryptedCredentials{},
		isRunner: true,
	}
}

func newClaudePermissionTestProvider() *mockAgentConfigProvider {
	agentfile := `
AGENT claude
EXECUTABLE claude
MODE pty
CONFIG permission_mode SELECT("default", "plan", "acceptEdits", "dontAsk", "bypassPermissions") = "bypassPermissions"
arg "--session-id" config.session_id when config.session_id != "" and not config.resume_enabled
arg "--resume" config.resume_session when config.resume_enabled
if config.permission_mode == "plan" and mode != "acp" {
  arg "--permission-mode" "plan"
}
if config.permission_mode != "default" and config.permission_mode != "plan" and config.permission_mode != "" {
  arg "--permission-mode" config.permission_mode
}
PROMPT_POSITION prepend
`
	return &mockAgentConfigProvider{
		agentDef: &agentDomain.Agent{
			Slug:              "claude-code",
			Name:              "Claude Code",
			LaunchCommand:     "claude",
			AdapterID:         "claude-stream-json",
			SupportedModes:    "pty",
			AgentfileSource:   &agentfile,
			UsesLegacyColumns: true,
		},
		config:   agentDomain.ConfigValues{},
		creds:    agentDomain.EncryptedCredentials{},
		isRunner: true,
	}
}

func setupOrchestrator(t *testing.T, opts ...func(*PodOrchestratorDeps)) (*PodOrchestrator, *PodService, *gorm.DB) {
	t.Helper()
	db := setupOrchestratorTestDB(t)
	podSvc := newTestPodService(db)

	provider := newTestProvider()
	configBuilder := agent.NewConfigBuilder(provider, noopBundleLoader{})

	deps := &PodOrchestratorDeps{
		PodService:    podSvc,
		ConfigBuilder: configBuilder,
		AgentResolver: &mockAgentResolver{agentDef: provider.agentDef},
		RunnerSelector: &mockRunnerSelector{
			resolveRunner: &runnerDomain.Runner{ID: 1},
		},
		ModelResources: &recordingModelResourceResolver{resource: resolvedResource("anthropic", "https://api.anthropic.com", "claude-test")},
	}

	for _, opt := range opts {
		opt(deps)
	}

	return NewPodOrchestrator(deps), podSvc, db
}

func withCoordinator(coord PodCoordinatorForOrchestrator) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.PodCoordinator = coord }
}

func withBilling(b BillingServiceForOrchestrator) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.BillingService = b }
}

func withUserSvc(u UserServiceForOrchestrator) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.UserService = u }
}

func withRepoSvc(r RepositoryServiceForOrchestrator) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.RepoService = r }
}

func withTicketSvc(ts TicketServiceForOrchestrator) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.TicketService = ts }
}

func withRunnerSelector(rs RunnerSelectorForOrchestrator) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.RunnerSelector = rs }
}

func withAgentResolver(ar AgentResolverForOrchestrator) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) { d.AgentResolver = ar }
}

func withAgentConfigProvider(provider *mockAgentConfigProvider) func(*PodOrchestratorDeps) {
	return func(d *PodOrchestratorDeps) {
		d.ConfigBuilder = agent.NewConfigBuilder(provider, noopBundleLoader{})
		d.AgentResolver = &mockAgentResolver{agentDef: provider.agentDef, err: provider.agentErr}
	}
}

func ptrStr(s string) *string { return &s }

func testModelResourceID() *int64 {
	id := int64(9)
	return &id
}
