package infra

import (
	"testing"
	"time"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/organization"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizeLockedApplyRequiresCurrentMembership(t *testing.T) {
	db, _ := orchestrationRepositoryForTest(t)
	state := controlservice.LockedApplyState{
		Plan: orchestrationTestCreatePlan(t),
	}

	err := authorizeLockedApply(db, orchestrationTestScope(), state)

	assert.ErrorIs(t, err, controlservice.ErrForbidden)
	require.NoError(t, db.Create(&organization.Member{
		OrganizationID: 42,
		UserID:         7,
		Role:           organization.RoleMember,
	}).Error)
	require.NoError(t, authorizeLockedApply(
		db,
		orchestrationTestScope(),
		state,
	))
}

func TestAuthorizeLockedApplyRequiresCreatorOrAdminForUpdate(t *testing.T) {
	db, _ := orchestrationRepositoryForTest(t)
	require.NoError(t, db.Create(&organization.Member{
		OrganizationID: 42,
		UserID:         7,
		Role:           organization.RoleMember,
	}).Error)
	head := orchestrationTestHead()
	head.CreatedByID = 8
	plan := orchestrationTestCreatePlan(t)
	plan.Operation = control.PlanOperationUpdate
	state := controlservice.LockedApplyState{Plan: plan, Head: &head}

	err := authorizeLockedApply(db, orchestrationTestScope(), state)

	assert.ErrorIs(t, err, controlservice.ErrForbidden)
	require.NoError(t, db.Model(&organization.Member{}).
		Where("organization_id = ? AND user_id = ?", 42, 7).
		Update("role", organization.RoleAdmin).Error)
	require.NoError(t, authorizeLockedApply(
		db,
		orchestrationTestScope(),
		state,
	))
}

func TestAuthorizeConsumedApplyRequiresCurrentMembership(t *testing.T) {
	db, _ := orchestrationRepositoryForTest(t)
	plan := orchestrationTestCreatePlan(t)
	head := orchestrationTestHead()
	applied, err := plan.Apply(
		plan.CreatedAt.Add(time.Minute),
		plan.ActorID,
		head.ID,
		head.Identity,
		head.ResourceVersion,
		head.Revision,
	)
	require.NoError(t, err)
	record, err := orchestrationPlanRecordFromDomain(applied)
	require.NoError(t, err)
	insertOrchestrationHead(t, db, head)
	require.NoError(t, db.Create(&record).Error)

	err = authorizeConsumedApply(
		db,
		orchestrationTestScope(),
		plan.ID,
	)

	assert.ErrorIs(t, err, controlservice.ErrForbidden)
	require.NoError(t, db.Create(&organization.Member{
		OrganizationID: 42,
		UserID:         7,
		Role:           organization.RoleMember,
	}).Error)
	require.NoError(t, authorizeConsumedApply(
		db,
		orchestrationTestScope(),
		plan.ID,
	))
}
