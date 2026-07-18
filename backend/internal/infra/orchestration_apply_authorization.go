package infra

import (
	"context"
	"errors"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/domain/organization"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type lockedApplyMemberReader struct {
	tx *gorm.DB
}

func (reader lockedApplyMemberReader) GetMember(
	ctx context.Context,
	organizationID int64,
	userID int64,
) (*organization.Member, error) {
	var member organization.Member
	err := reader.tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where(
			"organization_id = ? AND user_id = ?",
			organizationID,
			userID,
		).
		First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, organization.ErrMemberNotFound
	}
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func authorizeLockedApply(
	tx *gorm.DB,
	scope control.Scope,
	state controlservice.LockedApplyState,
) error {
	if tx == nil {
		return controlservice.ErrUnavailable
	}
	authorizer := controlservice.NewMemberAuthorizer(
		lockedApplyMemberReader{tx: tx},
	)
	if state.Plan.Operation == control.PlanOperationCreate {
		return authorizer.AuthorizeCreate(
			tx.Statement.Context,
			scope,
			state.Plan.Target,
		)
	}
	if state.Plan.Operation != control.PlanOperationUpdate || state.Head == nil {
		return control.ErrCorrupt
	}
	return authorizer.AuthorizeUpdate(
		tx.Statement.Context,
		scope,
		*state.Head,
	)
}

func authorizeConsumedApply(
	tx *gorm.DB,
	scope control.Scope,
	planID string,
) error {
	if tx == nil {
		return controlservice.ErrUnavailable
	}
	var planRecord orchestrationPlanRecord
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(
		"id = ? AND organization_id = ? AND actor_id = ?",
		planID,
		scope.OrganizationID,
		scope.ActorID,
	).First(&planRecord).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return control.ErrNotFound
	}
	if err != nil {
		return err
	}
	plan, err := planRecord.domain(scope)
	if err != nil {
		return err
	}
	if plan.Status != control.PlanStatusApplied ||
		plan.ResultIdentity == nil {
		return control.ErrConsumed
	}
	var headRecord orchestrationResourceRecord
	err = tx.Clauses(clause.Locking{Strength: "SHARE"}).
		Where(resourceIdentityQuery(scope, plan.Target)).
		First(&headRecord).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return control.ErrCorrupt
	}
	if err != nil {
		return err
	}
	head, err := headRecord.domain(scope)
	if err != nil {
		return err
	}
	if head.ID != plan.ResultResourceID ||
		head.Identity != *plan.ResultIdentity ||
		head.Revision != plan.ResultRevision ||
		head.ResourceVersion != plan.ResultResourceVersion {
		return control.ErrCorrupt
	}
	return authorizeLockedApply(
		tx,
		scope,
		controlservice.LockedApplyState{Plan: plan, Head: &head},
	)
}
