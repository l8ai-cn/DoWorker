package agentpod

import (
	"context"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareStructuredWorkerCreateReusesResolvedRepository(t *testing.T) {
	spec := podServiceWorkerSpec()
	repositoryID := int64(22)
	spec.Workspace.RepositoryID = &repositoryID
	spec.Workspace.Branch = "main"
	repository := &gitprovider.Repository{
		ID:             repositoryID,
		OrganizationID: 77,
		Slug:           "org/repo",
		HttpCloneURL:   "https://git.example.com/org/repo.git",
		IsActive:       true,
	}
	layer := "REPO \"org/repo\"\nBRANCH \"main\"\nMODE acp\n"
	preparer := &workerCreationPreparer{
		prepared: preparedWorkerSpecForArtifactTest(
			t,
			context.WithValue(
				context.WithValue(context.Background(), ctxKeyOrgID, int64(77)),
				ctxKeyUserID,
				int64(7),
			),
			spec,
			layer,
			repository,
		),
	}
	repositories := &mockRepoService{}
	orchestrator := NewPodOrchestrator(&PodOrchestratorDeps{
		WorkerCreation: preparer,
		RepoService:    repositories,
	})
	req := &OrchestrateCreatePodRequest{
		OrganizationID:  77,
		UserID:          7,
		WorkerSpecDraft: &workercreation.Draft{},
	}

	require.NoError(t, orchestrator.prepareStructuredWorkerCreate(context.Background(), req))
	require.NoError(t, orchestrator.preResolveFreshRepository(context.Background(), req, nil))

	require.Same(t, repository, req.preResolvedRepository)
	assert.Empty(t, repositories.getAccessibleCalls)
	assert.Empty(t, repositories.findAccessibleCalls)
}
