package agentpod

import (
	"context"
	"errors"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/user"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agent"
	userService "github.com/l8ai-cn/agentcloud/backend/internal/service/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserGitCredential_NilUserService(t *testing.T) {
	db := setupTestDB(t)
	orch := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:    newTestPodService(db),
		ConfigBuilder: agent.NewConfigBuilder(newTestProvider(), noopBundleLoader{}),
	})

	assert.Nil(t, orch.getUserGitCredential(context.Background(), 1))
}

func TestGetUserGitCredential_NoDefaultCredential(t *testing.T) {
	userSvc := &mockUserServiceForOrch{defaultCredErr: errors.New("not found")}
	db := setupTestDB(t)
	orch := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:    newTestPodService(db),
		ConfigBuilder: agent.NewConfigBuilder(newTestProvider(), noopBundleLoader{}),
		UserService:   userSvc,
	})

	assert.Nil(t, orch.getUserGitCredential(context.Background(), 1))
}

func TestGetUserGitCredential_RunnerLocal(t *testing.T) {
	userSvc := &mockUserServiceForOrch{
		defaultCred: &user.GitCredential{ID: 1, CredentialType: "runner_local"},
	}
	db := setupTestDB(t)
	orch := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:    newTestPodService(db),
		ConfigBuilder: agent.NewConfigBuilder(newTestProvider(), noopBundleLoader{}),
		UserService:   userSvc,
	})

	assert.Nil(t, orch.getUserGitCredential(context.Background(), 1))
}

func TestGetUserGitCredential_DecryptError(t *testing.T) {
	userSvc := &mockUserServiceForOrch{
		defaultCred:  &user.GitCredential{ID: 1, CredentialType: "oauth"},
		decryptedErr: errors.New("decrypt failed"),
	}
	db := setupTestDB(t)
	orch := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:    newTestPodService(db),
		ConfigBuilder: agent.NewConfigBuilder(newTestProvider(), noopBundleLoader{}),
		UserService:   userSvc,
	})

	assert.Nil(t, orch.getUserGitCredential(context.Background(), 1))
}

func TestGetUserGitCredential_SuccessPAT(t *testing.T) {
	userSvc := &mockUserServiceForOrch{
		defaultCred: &user.GitCredential{ID: 1, CredentialType: "pat"},
		decryptedCred: &userService.DecryptedCredential{
			Type: "pat", Token: "ghp_xxxxx",
		},
	}
	db := setupTestDB(t)
	orch := NewPodOrchestrator(&PodOrchestratorDeps{
		PodService:    newTestPodService(db),
		ConfigBuilder: agent.NewConfigBuilder(newTestProvider(), noopBundleLoader{}),
		UserService:   userSvc,
	})

	result := orch.getUserGitCredential(context.Background(), 1)
	require.NotNil(t, result)
	assert.Equal(t, "pat", result.Type)
	assert.Equal(t, "ghp_xxxxx", result.Token)
}

func TestCreatePod_RepoServiceErrorPropagates(t *testing.T) {
	repoErr := errors.New("repo not found")
	repoSvc := &mockRepoService{err: repoErr}
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord), withRepoSvc(repoSvc))
	repoID := int64(999)

	_, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("CONFIG mcp_enabled = true"),
		RepositoryID:    &repoID,
	})

	require.ErrorIs(t, err, repoErr)
	assert.False(t, coord.createPodCalled)
}

func TestBuildPodCommand_TicketServiceErrorIgnoresTicket(t *testing.T) {
	ticketSvc := &mockTicketServiceForOrch{err: errors.New("ticket not found")}
	coord := &mockPodCoordinator{}
	orch, _, _ := setupOrchestrator(t, withCoordinator(coord), withTicketSvc(ticketSvc))
	ticketID := int64(999)

	_, err := createPodWithPlanSourceForTest(t, orch, context.Background(), &OrchestrateCreatePodRequest{
		OrganizationID:  1,
		UserID:          1,
		RunnerID:        1,
		AgentSlug:       "claude-code",
		ModelResourceID: testModelResourceID(),
		AgentfileLayer:  ptrStr("CONFIG mcp_enabled = true"),
		TicketID:        &ticketID,
	})

	require.NoError(t, err)
}
