package agentpod

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	"github.com/anthropics/agentsmesh/backend/internal/domain/ticket"
	"github.com/anthropics/agentsmesh/backend/internal/domain/user"
	"github.com/anthropics/agentsmesh/backend/internal/service/agent"
	userService "github.com/anthropics/agentsmesh/backend/internal/service/user"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

// ==================== buildPodCommand Tests ====================

func TestBuildPodCommand_WithRepository(t *testing.T) {
	prepScript := "npm install"
	prepTimeout := 600
	repo := &gitprovider.Repository{
		HttpCloneURL:       "https://github.com/org/repo.git",
		DefaultBranch:      "develop",
		PreparationScript:  &prepScript,
		PreparationTimeout: &prepTimeout,
	}
	repoSvc := &mockRepoService{repo: repo}
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord), withRepoSvc(repoSvc))

	agentSlug := "claude-code"
	repoID := int64(10)
	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       agentSlug,
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("CONFIG mcp_enabled = true"),
		RepositoryID:    &repoID,
	})

	require.NoError(t, err)
	require.NotNil(t, coord.lastCmd)
	require.NotNil(t, coord.lastCmd.SandboxConfig)
	assert.Equal(t, "https://github.com/org/repo.git", coord.lastCmd.SandboxConfig.HttpCloneUrl)
	assert.Equal(t, "develop", coord.lastCmd.SandboxConfig.SourceBranch)
	assert.Equal(t, "npm install", coord.lastCmd.SandboxConfig.PreparationScript)
	assert.Equal(t, int32(600), coord.lastCmd.SandboxConfig.PreparationTimeout)
}

func TestBuildPodCommand_BranchOverride(t *testing.T) {
	repo := &gitprovider.Repository{
		HttpCloneURL:  "https://github.com/org/repo.git",
		DefaultBranch: "develop",
	}
	repoSvc := &mockRepoService{repo: repo}
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord), withRepoSvc(repoSvc))

	agentSlug := "claude-code"
	repoID := int64(10)
	branch := "feature/my-branch"
	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       agentSlug,
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("CONFIG mcp_enabled = true"),
		RepositoryID:    &repoID,
		BranchName:      &branch,
	})

	require.NoError(t, err)
	assert.Equal(t, "feature/my-branch", coord.lastCmd.SandboxConfig.SourceBranch)
}

func TestBuildPodCommand_WithTicket(t *testing.T) {
	ticketSvc := &mockTicketServiceForOrch{
		ticket: &ticket.Ticket{
			ID:   1,
			Slug: "AM-42",
		},
	}
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord), withTicketSvc(ticketSvc))

	agentSlug := "claude-code"
	ticketID := int64(1)
	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       agentSlug,
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("CONFIG mcp_enabled = true"),
		TicketID:        &ticketID,
	})

	require.NoError(t, err)
	assert.True(t, coord.createPodCalled)
}

func TestBuildPodCommand_WithTicketSlug(t *testing.T) {
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord))

	agentSlug := "claude-code"
	ticketSlug := "AM-99"
	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       agentSlug,
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("CONFIG mcp_enabled = true"),
		TicketSlug:      &ticketSlug,
	})

	require.NoError(t, err)
	assert.True(t, coord.createPodCalled)
}

func TestBuildPodCommand_GeminiUsesExactModelResource(t *testing.T) {
	agentfileSource := "AGENT gemini\nEXECUTABLE gemini\nMODE pty\nENV GEMINI_API_KEY SECRET OPTIONAL\nPROMPT_POSITION append\n"
	provider := &mockAgentConfigProvider{
		agentDef: &agentDomain.Agent{
			Slug:              "gemini-cli",
			LaunchCommand:     "gemini",
			AdapterID:         "gemini-acp",
			SupportedModes:    "pty",
			AgentfileSource:   &agentfileSource,
			UsesLegacyColumns: false,
		},
	}
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t,
		withCoordinator(coord),
		withAgentConfigProvider(provider),
		func(deps *PodOrchestratorDeps) {
			deps.ModelResources = &recordingModelResourceResolver{
				resource: resolvedResource("gemini", "", "gemini-pro"),
			}
		},
	)

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "gemini-cli",
		ModelResourceID: testModelResourceID(),
		AutomationLevel: podDomain.AutomationLevelInteractive,
	})

	require.NoError(t, err)
	require.NotNil(t, coord.lastCmd)
	assert.Equal(t, "sk-test", coord.lastCmd.EnvVars["GEMINI_API_KEY"])
	assert.Equal(t, []string{"--model", "gemini-pro"}, coord.lastCmd.LaunchArgs)
}

func TestBuildPodCommand_MiniMaxPlacesModelBeforeMessage(t *testing.T) {
	agentfileSource := "AGENT mmx\nEXECUTABLE mmx\nMODE pty\nCONFIG model STRING = \"\"\nCONFIG base_url STRING = \"\"\nENV MINIMAX_API_KEY SECRET OPTIONAL\nPROMPT_POSITION append\narg \"text\"\narg \"chat\"\narg \"--model\" config.model when config.model != \"\"\narg \"--base-url\" config.base_url when config.base_url != \"\"\narg \"--message\"\n"
	provider := &mockAgentConfigProvider{
		agentDef: &agentDomain.Agent{
			Slug:              "minimax-cli",
			LaunchCommand:     "mmx",
			Executable:        "mmx",
			AdapterID:         "minimax-pty",
			SupportedModes:    "pty",
			AgentfileSource:   &agentfileSource,
			UsesLegacyColumns: false,
		},
	}
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t,
		withCoordinator(coord),
		withAgentConfigProvider(provider),
		func(deps *PodOrchestratorDeps) {
			deps.ModelResources = &recordingModelResourceResolver{
				resource: resolvedResource("minimax", "https://api.minimax.io/v1", "MiniMax-M2.5"),
			}
		},
	)

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "minimax-cli",
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr(`PROMPT "Reply with exactly READY."`),
		AutomationLevel: podDomain.AutomationLevelInteractive,
	})

	require.NoError(t, err)
	require.NotNil(t, coord.lastCmd)
	assert.Equal(t, "sk-test", coord.lastCmd.EnvVars["MINIMAX_API_KEY"])
	assert.Equal(t, "Reply with exactly READY.", coord.lastCmd.Prompt)
	assert.Equal(t, "append", coord.lastCmd.PromptPosition)
	assert.Equal(t, []string{"text", "chat", "--model", "MiniMax-M2.5", "--base-url", "https://api.minimax.io", "--message"}, coord.lastCmd.LaunchArgs)
}

func TestBuildPodCommand_PreservesUserPromptForHermes(t *testing.T) {
	agentfileSource := "AGENT hermes\nEXECUTABLE hermes\nMODE pty\nPROMPT_POSITION after_first\nCONFIG provider = \"openai\"\nCONFIG model = \"gpt-5\"\narg \"--oneshot\"\narg \"--provider\" config.provider\narg \"--model\" config.model\n"
	provider := &mockAgentConfigProvider{
		agentDef: &agentDomain.Agent{
			Slug: "hermes", LaunchCommand: "hermes", AdapterID: "hermes-pty", SupportedModes: "pty",
			AgentfileSource: &agentfileSource,
		},
	}
	spec := podServiceWorkerSpec()
	spec.Runtime.WorkerType.Slug = slugkit.MustNewForTest("hermes")
	spec.TypeConfig.InteractionMode = "pty"
	spec.Workspace.InitialTask = "Reply with exactly READY."
	resolved := resolvedWorkerSpecFromSpecForPodServiceTest(t, 1, spec)
	preparer := &workerCreationPreparer{
		prepared: workercreation.Prepared{
			Snapshot:       resolved,
			Spec:           spec,
			AgentfileLayer: "MODE pty\nPROMPT \"Reply with exactly READY.\"\n",
		},
	}
	resource := resolvedOpenAIResource()
	resource.Connection.ID = spec.Runtime.ModelBinding.ConnectionID
	resource.Connection.Revision = spec.Runtime.ModelBinding.ConnectionRevision
	resource.Connection.ProviderKey = spec.Runtime.ModelBinding.ProviderKey
	resource.Resource.ID = spec.Runtime.ModelBinding.ResourceID
	resource.Resource.ProviderConnectionID = resource.Connection.ID
	resource.Resource.Revision = spec.Runtime.ModelBinding.ResourceRevision
	resource.Resource.ModelID = spec.Runtime.ModelBinding.ModelID
	coord := &mockPodCoordinator{}
	orch, _, db := setupOrchestrator(t,
		withCoordinator(coord),
		withAgentConfigProvider(provider),
		func(deps *PodOrchestratorDeps) {
			deps.ModelResources = &recordingModelResourceResolver{resource: resource}
			deps.WorkerCreation = preparer
		},
	)
	require.NoError(t, db.Exec(`CREATE TABLE worker_spec_snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		organization_id INTEGER NOT NULL,
		version INTEGER NOT NULL,
		spec_json BLOB NOT NULL,
		summary_json BLOB NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`).Error)

	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		WorkerSpecDraft: &workercreation.Draft{},
	})

	require.NoError(t, err)
	require.NotNil(t, coord.lastCmd)
	assert.Equal(t, "Reply with exactly READY.", coord.lastCmd.Prompt)
	assert.Equal(t, "after_first", coord.lastCmd.PromptPosition)
	assert.Equal(t, []string{"--oneshot", "--provider", "openai", "--model", "gpt-5"}, coord.lastCmd.LaunchArgs)
}

func TestBuildPodCommand_WithOAuthCredential(t *testing.T) {
	userSvc := &mockUserServiceForOrch{
		defaultCred: &user.GitCredential{
			ID:             1,
			CredentialType: "oauth",
		},
		decryptedCred: &userService.DecryptedCredential{
			Type:  "oauth",
			Token: "github-token-123",
		},
	}
	repo := &gitprovider.Repository{
		HttpCloneURL: "https://github.com/org/repo.git",
	}
	repoSvc := &mockRepoService{repo: repo}
	coord := &mockPodCoordinator{}
	repoID := int64(10)
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord), withUserSvc(userSvc), withRepoSvc(repoSvc))

	agentSlug := "claude-code"
	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       agentSlug,
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("CONFIG mcp_enabled = true"),
		RepositoryID:    &repoID,
	})

	require.NoError(t, err)
	require.NotNil(t, coord.lastCmd.SandboxConfig)
	assert.Equal(t, "oauth", coord.lastCmd.SandboxConfig.CredentialType)
	assert.Equal(t, "github-token-123", coord.lastCmd.SandboxConfig.GitToken)
}

func TestBuildPodCommand_WithSSHCredential(t *testing.T) {
	userSvc := &mockUserServiceForOrch{
		defaultCred: &user.GitCredential{
			ID:             1,
			CredentialType: "ssh_key",
		},
		decryptedCred: &userService.DecryptedCredential{
			Type:          "ssh_key",
			SSHPrivateKey: "-----BEGIN RSA PRIVATE KEY-----\nfake\n-----END RSA PRIVATE KEY-----",
		},
	}
	repo := &gitprovider.Repository{
		SshCloneURL: "git@github.com:org/repo.git",
	}
	repoSvc := &mockRepoService{repo: repo}
	coord := &mockPodCoordinator{}
	repoID := int64(10)
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord), withUserSvc(userSvc), withRepoSvc(repoSvc))

	agentSlug := "claude-code"
	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       agentSlug,
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("CONFIG mcp_enabled = true"),
		RepositoryID:    &repoID,
	})

	require.NoError(t, err)
	require.NotNil(t, coord.lastCmd.SandboxConfig)
	assert.Equal(t, "ssh_key", coord.lastCmd.SandboxConfig.CredentialType)
	assert.Contains(t, coord.lastCmd.SandboxConfig.SshPrivateKey, "BEGIN RSA PRIVATE KEY")
}

func TestBuildPodCommand_RunnerLocalCredential_NoCredsSent(t *testing.T) {
	userSvc := &mockUserServiceForOrch{
		defaultCred: &user.GitCredential{
			ID:             1,
			CredentialType: "runner_local",
		},
	}
	repo := &gitprovider.Repository{
		HttpCloneURL: "https://github.com/org/repo.git",
	}
	repoSvc := &mockRepoService{repo: repo}
	coord := &mockPodCoordinator{}
	repoID := int64(10)
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord), withUserSvc(userSvc), withRepoSvc(repoSvc))

	agentSlug := "claude-code"
	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       agentSlug,
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("CONFIG mcp_enabled = true"),
		RepositoryID:    &repoID,
	})

	require.NoError(t, err)
	require.NotNil(t, coord.lastCmd.SandboxConfig)
	assert.Empty(t, coord.lastCmd.SandboxConfig.CredentialType)
	assert.Empty(t, coord.lastCmd.SandboxConfig.GitToken)
}

// ==================== getUserGitCredential Tests ====================

func TestGetUserGitCredential_NilUserService(t *testing.T) {
	db := setupTestDB(t)
	podSvc := newTestPodService(db)
	provider := newTestProvider()
	orch := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:    podSvc,
		ConfigBuilder: agent.NewConfigBuilder(provider, noopBundleLoader{}),
	})

	result := orch.getUserGitCredential(context.Background(), 1)
	assert.Nil(t, result)
}

func TestGetUserGitCredential_NoDefaultCredential(t *testing.T) {
	userSvc := &mockUserServiceForOrch{
		defaultCred:    nil,
		defaultCredErr: errors.New("not found"),
	}
	db := setupTestDB(t)
	podSvc := newTestPodService(db)
	provider := newTestProvider()
	orch := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:    podSvc,
		ConfigBuilder: agent.NewConfigBuilder(provider, noopBundleLoader{}),
		UserService:   userSvc,
	})

	result := orch.getUserGitCredential(context.Background(), 1)
	assert.Nil(t, result)
}

func TestGetUserGitCredential_RunnerLocal(t *testing.T) {
	userSvc := &mockUserServiceForOrch{
		defaultCred: &user.GitCredential{
			ID:             1,
			CredentialType: "runner_local",
		},
	}
	db := setupTestDB(t)
	podSvc := newTestPodService(db)
	provider := newTestProvider()
	orch := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:    podSvc,
		ConfigBuilder: agent.NewConfigBuilder(provider, noopBundleLoader{}),
		UserService:   userSvc,
	})

	result := orch.getUserGitCredential(context.Background(), 1)
	assert.Nil(t, result) // runner_local returns nil
}

func TestGetUserGitCredential_DecryptError(t *testing.T) {
	userSvc := &mockUserServiceForOrch{
		defaultCred: &user.GitCredential{
			ID:             1,
			CredentialType: "oauth",
		},
		decryptedErr: errors.New("decrypt failed"),
	}
	db := setupTestDB(t)
	podSvc := newTestPodService(db)
	provider := newTestProvider()
	orch := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:    podSvc,
		ConfigBuilder: agent.NewConfigBuilder(provider, noopBundleLoader{}),
		UserService:   userSvc,
	})

	result := orch.getUserGitCredential(context.Background(), 1)
	assert.Nil(t, result) // Error during decrypt -> returns nil
}

func TestGetUserGitCredential_Success_PAT(t *testing.T) {
	userSvc := &mockUserServiceForOrch{
		defaultCred: &user.GitCredential{
			ID:             1,
			CredentialType: "pat",
		},
		decryptedCred: &userService.DecryptedCredential{
			Type:  "pat",
			Token: "ghp_xxxxx",
		},
	}
	db := setupTestDB(t)
	podSvc := newTestPodService(db)
	provider := newTestProvider()
	orch := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:    podSvc,
		ConfigBuilder: agent.NewConfigBuilder(provider, noopBundleLoader{}),
		UserService:   userSvc,
	})

	result := orch.getUserGitCredential(context.Background(), 1)
	require.NotNil(t, result)
	assert.Equal(t, "pat", result.Type)
	assert.Equal(t, "ghp_xxxxx", result.Token)
}

// ==================== Service Error Tests ====================

func TestCreatePod_RepoServiceError_Propagates(t *testing.T) {
	repoErr := errors.New("repo not found")
	repoSvc := &mockRepoService{err: repoErr}
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord), withRepoSvc(repoSvc))

	agentSlug := "claude-code"
	repoID := int64(999)
	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       agentSlug,
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("CONFIG mcp_enabled = true"),
		RepositoryID:    &repoID,
	})

	require.ErrorIs(t, err, repoErr)
	assert.False(t, coord.createPodCalled)
}

func TestBuildPodCommand_TicketServiceError_IgnoresTicket(t *testing.T) {
	ticketSvc := &mockTicketServiceForOrch{err: errors.New("ticket not found")}
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord), withTicketSvc(ticketSvc))

	agentSlug := "claude-code"
	ticketID := int64(999)
	_, err := orch.CreatePod(context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       agentSlug,
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("CONFIG mcp_enabled = true"),
		TicketID:        &ticketID,
	})

	require.NoError(t, err) // Ticket error is not fatal
}
