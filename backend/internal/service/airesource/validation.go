package airesource

import (
	"context"
	"errors"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/pkg/audit"
)

func (s *Service) ValidateConnection(ctx context.Context, actor Actor, connectionID int64) error {
	connection, _, err := s.connectionForActor(ctx, actor, connectionID, true)
	if err != nil {
		return err
	}
	provider, exists := domain.Provider(connection.ProviderKey.String())
	if !exists {
		return ErrInvalidProvider
	}
	startedAt := time.Now().UTC()
	connection.Status = domain.ConnectionStatusUnchecked
	connection.LastValidatedAt = &startedAt
	connection.ValidationError = "validation in progress"
	if err := s.persistValidationState(ctx, actor, connection, audit.ActionProviderConnectionValidationStarted, "started"); err != nil {
		return err
	}
	credentials, validationErr := s.decryptCredentials(connection)
	if validationErr == nil {
		validationErr = s.endpoints.Validate(ctx, connection.BaseURL)
	}
	if validationErr == nil {
		validationErr = s.prober.Probe(ctx, ProbeInput{Provider: provider, BaseURL: connection.BaseURL, Credentials: credentials})
	}
	finishedAt := time.Now().UTC()
	connection.LastValidatedAt = &finishedAt
	connection.ValidationError = safeValidationError(validationErr)
	connection.Status = domain.ConnectionStatusValid
	if errors.Is(validationErr, ErrProbeUnsupported) {
		connection.Status = domain.ConnectionStatusUnchecked
	} else if validationErr != nil {
		connection.Status = domain.ConnectionStatusInvalid
	}
	mutationErr := s.persistValidationState(ctx, actor, connection, audit.ActionProviderConnectionValidated, validationResult(validationErr))
	if mutationErr != nil {
		if validationErr != nil && errors.Is(mutationErr, ErrAudit) {
			return auditValidationError(validationErr, mutationErr)
		}
		return mutationErr
	}
	if validationErr != nil {
		result := auditValidationError(validationErr, nil)
		if errors.Is(validationErr, ErrDecrypt) || errors.Is(validationErr, ErrInvalidCredentials) || errors.Is(validationErr, ErrInvalidEndpoint) || errors.Is(validationErr, ErrProbeUnsupported) {
			result = errors.Join(result, validationErr)
		}
		return result
	}
	return nil
}

func (s *Service) persistValidationState(ctx context.Context, actor Actor, connection *domain.Connection, action, result string) error {
	return s.mutations.Run(ctx, func(repo domain.Repository, recorder AuditRecorder) error {
		revision, err := repo.SetValidationState(
			ctx,
			connection.ID,
			connection.Revision,
			connection.CredentialsEncrypted,
			connection.Status,
			*connection.LastValidatedAt,
			connection.ValidationError,
		)
		if err != nil {
			return err
		}
		connection.Revision = revision
		return recordAudit(ctx, recorder, actor, action, audit.ResourceProviderConnection, connection.ID, connection, result, audit.Details{"status": string(connection.Status)})
	})
}
