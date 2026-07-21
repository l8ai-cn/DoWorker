package agentpod

import (
	"errors"
	"reflect"
	"strings"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/gitprovider"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	workercreation "github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
)

func validatePreparedWorkerSnapshot(
	snapshot specdomain.Snapshot,
	prepared workercreation.PreparedSnapshot,
) error {
	spec, err := specdomain.NormalizeAndValidate(snapshot.Spec)
	if err != nil {
		return errors.Join(ErrWorkerSpecSnapshotMismatch, err)
	}
	if strings.TrimSpace(prepared.AgentfileLayer) == "" ||
		!reflect.DeepEqual(spec, prepared.Spec) {
		return ErrInvalidPreparedWorkerSpec
	}
	return validatePreparedWorkerRepository(
		snapshot.OrganizationID,
		prepared.Spec.Workspace.RepositoryID,
		prepared.Repository,
	)
}

func validatePreparedWorkerRepository(
	organizationID int64,
	repositoryID *int64,
	repository *gitprovider.Repository,
) error {
	if repositoryID == nil {
		if repository != nil {
			return ErrInvalidPreparedWorkerSpec
		}
		return nil
	}
	if repository == nil ||
		repository.ID != *repositoryID ||
		repository.OrganizationID != organizationID ||
		!repository.IsActive {
		return ErrInvalidPreparedWorkerSpec
	}
	return nil
}
