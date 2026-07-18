package airesource

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/pkg/audit"
)

func recordAudit(ctx context.Context, recorder AuditRecorder, actor Actor, action, resourceType string, resourceID int64, connection *domain.Connection, result string, details audit.Details) error {
	actorID := actor.UserID
	if details == nil {
		details = audit.Details{}
	}
	details["correlation_id"] = actor.CorrelationID
	details["owner_scope"] = string(connection.OwnerScope)
	details["owner_id"] = connection.OwnerID
	details["provider_key"] = connection.ProviderKey.String()
	details["result"] = result
	entry := audit.Entry(action).Actor(audit.ActorTypeUser, &actorID).Resource(resourceType, &resourceID).Details(details)
	if connection.OwnerScope == domain.OwnerScopeOrg {
		entry.Organization(connection.OwnerID)
	}
	if err := recorder.Record(ctx, entry.Build()); err != nil {
		return errors.Join(ErrAudit, err)
	}
	return nil
}

func validationResult(err error) string {
	if err == nil {
		return "success"
	}
	return "failure"
}

func safeValidationError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, ErrInvalidCredentials) {
		return "credentials rejected"
	}
	if errors.Is(err, ErrInvalidEndpoint) {
		return "endpoint rejected"
	}
	if errors.Is(err, ErrProbeUnsupported) {
		return "validation unavailable"
	}
	if errors.Is(err, ErrProviderEndpointUnavailable) {
		return "provider endpoint unavailable"
	}
	return "connection validation failed"
}

func auditValidationError(probeErr, auditErr error) error {
	validationErr := fmt.Errorf("%w: %s", ErrValidation, safeValidationError(probeErr))
	if auditErr != nil {
		return errors.Join(validationErr, auditErr)
	}
	return validationErr
}
