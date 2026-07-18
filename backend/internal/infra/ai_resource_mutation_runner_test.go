package infra

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	service "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/anthropics/agentsmesh/backend/pkg/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAIResourceMutationRunnerSharesTransactionForRepositoryAndAudit(t *testing.T) {
	db, _ := setupAIResourceRepository(t)
	runner := NewAIResourceMutationRunner(db)
	ctx := context.Background()
	err := runner.Run(ctx, func(repo domain.Repository, recorder service.AuditRecorder) error {
		connection := validAIConnection(domain.OwnerScopeUser, 1, "rolled-back", true)
		require.NoError(t, repo.CreateConnection(ctx, connection))
		require.NoError(t, recorder.Record(ctx, audit.Entry(audit.ActionProviderConnectionCreated).Organization(10).Actor(audit.ActorTypeUser, nil).Resource(audit.ResourceProviderConnection, &connection.ID).Build()))
		return errors.New("rollback")
	})
	require.Error(t, err)
	var connectionCount, auditCount int64
	require.NoError(t, db.Table("provider_connections").Where("identifier = ?", "rolled-back").Count(&connectionCount).Error)
	require.NoError(t, db.Table("audit_logs").Where("action = ?", audit.ActionProviderConnectionCreated).Count(&auditCount).Error)
	assert.Zero(t, connectionCount)
	assert.Zero(t, auditCount)

	require.NoError(t, runner.Run(ctx, func(repo domain.Repository, recorder service.AuditRecorder) error {
		connection := validAIConnection(domain.OwnerScopeUser, 1, "committed", true)
		if createErr := repo.CreateConnection(ctx, connection); createErr != nil {
			return createErr
		}
		return recorder.Record(ctx, audit.Entry(audit.ActionProviderConnectionCreated).Organization(10).Actor(audit.ActorTypeUser, nil).Resource(audit.ResourceProviderConnection, &connection.ID).Build())
	}))
	require.NoError(t, db.Table("provider_connections").Where("identifier = ?", "committed").Count(&connectionCount).Error)
	require.NoError(t, db.Table("audit_logs").Where("action = ?", audit.ActionProviderConnectionCreated).Count(&auditCount).Error)
	assert.EqualValues(t, 1, connectionCount)
	assert.EqualValues(t, 1, auditCount)
}

func TestAIResourceMutationRunnerRollsBackNestedValidationWhenAuditFails(t *testing.T) {
	db, repo := setupAIResourceRepository(t)
	connection := createAIConnection(t, repo, domain.OwnerScopeUser, 1, "openai-main", true)
	resource := createAIResource(t, repo, connection.ID, "chat-model", true, domain.ModalityChat)
	require.NoError(t, db.Exec(`CREATE TRIGGER reject_ai_resource_audit BEFORE INSERT ON audit_logs BEGIN SELECT RAISE(FAIL, 'injected'); END`).Error)
	runner := NewAIResourceMutationRunner(db)
	err := runner.Run(context.Background(), func(txRepo domain.Repository, recorder service.AuditRecorder) error {
		if _, stateErr := txRepo.SetValidationState(context.Background(), connection.ID, connection.Revision, connection.CredentialsEncrypted, domain.ConnectionStatusInvalid, time.Now(), "credentials rejected"); stateErr != nil {
			return stateErr
		}
		return recorder.Record(context.Background(), audit.Entry(audit.ActionProviderConnectionValidated).Organization(10).Actor(audit.ActorTypeUser, nil).Resource(audit.ResourceProviderConnection, &connection.ID).Build())
	})
	require.Error(t, err)
	storedConnection, loadErr := repo.GetConnectionByID(context.Background(), connection.ID)
	require.NoError(t, loadErr)
	storedResource, loadErr := repo.GetResourceByID(context.Background(), resource.ID)
	require.NoError(t, loadErr)
	assert.Equal(t, domain.ConnectionStatusValid, storedConnection.Status)
	assert.Equal(t, domain.ConnectionStatusValid, storedResource.Status)
}
