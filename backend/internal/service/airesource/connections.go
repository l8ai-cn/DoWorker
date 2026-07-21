package airesource

import (
	"context"
	"fmt"
	"strings"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/audit"
)

func (s *Service) CreateConnection(ctx context.Context, actor Actor, input CreateConnectionInput) (ConnectionView, error) {
	if _, err := s.authorizeOwner(ctx, actor, input.OwnerScope, input.OwnerID, true); err != nil {
		return ConnectionView{}, err
	}
	provider, exists := domain.Provider(input.ProviderKey.String())
	if !exists {
		return ConnectionView{}, ErrInvalidProvider
	}
	baseURL, err := s.validatedBaseURL(ctx, provider, input.BaseURL)
	if err != nil {
		return ConnectionView{}, err
	}
	encrypted, configured, err := s.encryptCredentials(provider, input.Credentials)
	if err != nil {
		return ConnectionView{}, err
	}
	connection := &domain.Connection{
		OwnerScope: input.OwnerScope, OwnerID: input.OwnerID, Identifier: input.Identifier,
		ProviderKey: input.ProviderKey, Name: strings.TrimSpace(input.Name), BaseURL: baseURL,
		CredentialsEncrypted: encrypted, ConfiguredFields: configured, Status: domain.ConnectionStatusUnchecked,
		IsEnabled: true, CreatedBy: actor.UserID,
	}
	if err := connection.ValidateIdentifiers(); err != nil {
		return ConnectionView{}, fmt.Errorf("%w: %v", ErrInvalidOwner, err)
	}
	err = s.mutations.Run(ctx, func(repo domain.Repository, recorder AuditRecorder) error {
		if createErr := repo.CreateConnection(ctx, connection); createErr != nil {
			return createErr
		}
		return recordAudit(ctx, recorder, actor, audit.ActionProviderConnectionCreated, audit.ResourceProviderConnection, connection.ID, connection, "success", nil)
	})
	if err != nil {
		return ConnectionView{}, err
	}
	view := connectionView(connection, true, nil)
	return view, nil
}

func (s *Service) UpdateConnection(ctx context.Context, actor Actor, connectionID int64, input UpdateConnectionInput) (ConnectionView, error) {
	connection, _, err := s.connectionForActor(ctx, actor, connectionID, true)
	if err != nil {
		return ConnectionView{}, err
	}
	provider, exists := domain.Provider(connection.ProviderKey.String())
	if !exists {
		return ConnectionView{}, ErrInvalidProvider
	}
	if strings.TrimSpace(input.Name) != "" {
		connection.Name = strings.TrimSpace(input.Name)
	}
	resetValidation := false
	runtimeChanged := false
	if strings.TrimSpace(input.BaseURL) != "" {
		baseURL, validationErr := s.validatedBaseURL(ctx, provider, input.BaseURL)
		if validationErr != nil {
			return ConnectionView{}, validationErr
		}
		if baseURL != connection.BaseURL {
			connection.BaseURL = baseURL
			runtimeChanged = true
			resetValidation = true
		}
	}
	if input.Credentials != nil {
		connection.CredentialsEncrypted, connection.ConfiguredFields, err = s.encryptCredentials(provider, input.Credentials)
		if err != nil {
			return ConnectionView{}, err
		}
		resetValidation = true
	}
	if resetValidation {
		markConnectionUnchecked(connection)
	}
	err = s.mutations.Run(ctx, func(repo domain.Repository, recorder AuditRecorder) error {
		save := repo.SaveConnectionMetadata
		if runtimeChanged {
			save = repo.SaveConnection
		}
		if saveErr := save(ctx, connection); saveErr != nil {
			return saveErr
		}
		return recordAudit(ctx, recorder, actor, audit.ActionProviderConnectionUpdated, audit.ResourceProviderConnection, connection.ID, connection, "success", nil)
	})
	if err != nil {
		return ConnectionView{}, err
	}
	view := connectionView(connection, true, nil)
	return view, nil
}

func (s *Service) RotateConnectionCredentials(ctx context.Context, actor Actor, connectionID int64, credentials map[string]string) error {
	connection, _, err := s.connectionForActor(ctx, actor, connectionID, true)
	if err != nil {
		return err
	}
	provider, exists := domain.Provider(connection.ProviderKey.String())
	if !exists {
		return ErrInvalidProvider
	}
	connection.CredentialsEncrypted, connection.ConfiguredFields, err = s.encryptCredentials(provider, credentials)
	if err != nil {
		return err
	}
	markConnectionUnchecked(connection)
	return s.mutations.Run(ctx, func(repo domain.Repository, recorder AuditRecorder) error {
		if saveErr := repo.SaveConnectionMetadata(ctx, connection); saveErr != nil {
			return saveErr
		}
		return recordAudit(ctx, recorder, actor, audit.ActionProviderConnectionCredentialsRotated, audit.ResourceProviderConnection, connection.ID, connection, "success", nil)
	})
}

func (s *Service) SetConnectionEnabled(ctx context.Context, actor Actor, connectionID int64, enabled bool) error {
	connection, _, err := s.connectionForActor(ctx, actor, connectionID, true)
	if err != nil {
		return err
	}
	connection.IsEnabled = enabled
	action := audit.ActionProviderConnectionDisabled
	if enabled {
		action = audit.ActionProviderConnectionEnabled
	}
	return s.mutations.Run(ctx, func(repo domain.Repository, recorder AuditRecorder) error {
		if saveErr := repo.SaveConnectionMetadata(ctx, connection); saveErr != nil {
			return saveErr
		}
		return recordAudit(ctx, recorder, actor, action, audit.ResourceProviderConnection, connection.ID, connection, "success", nil)
	})
}

func (s *Service) DeleteConnection(ctx context.Context, actor Actor, connectionID int64) error {
	connection, _, err := s.connectionForActor(ctx, actor, connectionID, true)
	if err != nil {
		return err
	}
	return s.mutations.Run(ctx, func(repo domain.Repository, recorder AuditRecorder) error {
		if deleteErr := repo.DeleteConnection(ctx, connectionID, connection.Revision, connection.UpdatedAt); deleteErr != nil {
			return deleteErr
		}
		return recordAudit(ctx, recorder, actor, audit.ActionProviderConnectionDeleted, audit.ResourceProviderConnection, connection.ID, connection, "success", nil)
	})
}

func markConnectionUnchecked(connection *domain.Connection) {
	connection.Status = domain.ConnectionStatusUnchecked
	connection.LastValidatedAt = nil
	connection.ValidationError = ""
}
