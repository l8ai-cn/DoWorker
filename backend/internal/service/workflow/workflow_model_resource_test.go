package workflow

import (
	"context"
	"testing"

	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/stretchr/testify/require"
)

func TestWorkflowServicePersistsModelResourceID(t *testing.T) {
	db := testkit.SetupTestDB(t)
	svc := NewWorkflowService(infra.NewWorkflowRepository(db))
	ctx := context.Background()
	resourceID := int64(42)

	created, err := svc.Create(ctx, &CreateWorkflowRequest{
		OrganizationID:  1,
		CreatedByID:     1,
		Name:            "Nightly",
		Slug:            "nightly",
		AgentSlug:       "claude-code",
		PromptTemplate:  "run tests",
		ModelResourceID: &resourceID,
		ExecutionMode:   workflowDomain.ExecutionModeDirect,
		SandboxStrategy: workflowDomain.SandboxStrategyFresh,
		TimeoutMinutes:  30,
		IdleTimeoutSec:  30,
	})
	require.NoError(t, err)
	require.Equal(t, &resourceID, created.ModelResourceID)

	replacementID := int64(77)
	updated, err := svc.Update(ctx, 1, "nightly", &UpdateWorkflowRequest{
		ModelResourceID: &replacementID,
	})
	require.NoError(t, err)
	require.Equal(t, &replacementID, updated.ModelResourceID)
}

func TestBuildLoopCreatePodRequestCarriesModelResourceID(t *testing.T) {
	resourceID := int64(42)
	workflow := &workflowDomain.Workflow{
		OrganizationID:     1,
		AgentSlug:          "claude-code",
		ModelResourceID:    &resourceID,
		PromptTemplate:     "run tests",
		PermissionMode:     "bypassPermissions",
		SessionPersistence: true,
	}

	req := buildWorkflowCreatePodRequest(workflow, 9, "PROMPT \"run tests\"", "", true)

	require.Equal(t, &resourceID, req.ModelResourceID)
	require.NotContains(t, *req.AgentfileLayer, "credential")
}
